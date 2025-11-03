# Algorithm analysis

## Fixed Window Algorithm

### Mental Model

This algorithm works by dividing time into equal parts (called *windows*).
Each of these windows has a predefined limit of requests.
Once the limit is reached, any further requests within that window are rejected.

For example, if the current time is `12:24:02` and the window size is one minute, the time is truncated to `12:24:00`.
This means all requests between `12:24:00` and `12:24:59` belong to the same window.

However, this leads to **burst errors**. For example, if someone sends 10 requests until `12:24:59`, they can again send 10 requests at `12:25:00`, effectively making 20 requests within a 2-second period.

### How I Implemented It

I implemented this by taking the current time and truncating it to the window size (for example, one minute).
The truncated time is used to identify the active window.

If the storage type is Redis:

  * I used **Lua scripts**, as they allow atomic operations (`GET`, `SET`, and `UPDATE` in a single step).
  * Steps:
    1.  Get the current window based on local time.
    2.  Check if `(current_window_requests + increment) > limit`.
    3.  If yes → reject the request.
    4.  Else → increment the counter and set an expiration time for the key.

If the storage type is in-memory:

  * I used **`sync.Map`** because, as mentioned in Go’s documentation, when a given key is written once but read many times (such as in caches that only grow), it performs better than a regular map with manual locking.
  * Steps:
    1.  Check if the data exists.
    2.  If not, create a new entry with an initial value and expiration.
    3.  Check if the existing entry has expired.
    4.  If expired, replace it with a new entry and updated timestamp (using compare-and-swap).
    5.  Otherwise, increment the value safely using compare-and-swap.

### Edge Cases I Discovered

1.  **Burst at window boundary:**

      * Problem: Users can exceed the effective rate limit by sending requests near the window edge.
      * Solution: Use the **Sliding Window Counter** algorithm to smooth out the rate.

2.  **Clock drift:**

      * Problem: In distributed systems, clocks on different servers may vary, causing inconsistent window boundaries.
      * Solution: Use Redis time (`TIME` command) as the single source of truth.

### When to Use

  * Suitable for **simple** and **low-traffic** systems.
  * Works well in **single-node** setups.
  * Not recommended for **distributed** environments where accurate rate enforcement is critical.

### Performance Characteristics

  * Constant time (**O(1)**) for each operation.
  * **In-Memory (Storage Benchmark):** 389.8 ns/op (263 B/op, 7 allocs/op).
  * **In-Memory (Single Key, Concurrent):** 1055 ns/op (393 B/op, 12 allocs/op).
  * **In-Memory (Multiple Keys, Concurrent):** 104.0 ns/op (262 B/op, 7 allocs/op).
  * **Redis (Storage Benchmark):** 23354 ns/op (1352 B/op, 25 allocs/op).
  * May experience burst traffic at window edges.

-----

## Token Bucket Algorithm

### Mental Model

This algorithm assigns each user a *bucket* (a counter).
Each bucket is initially filled with a fixed number of tokens (equal to the rate limit).
Every request consumes one token. If the number of tokens is greater than zero, the request is allowed.
Tokens are refilled at a constant rate (for example, 10 tokens per second).

Since refilling can happen at fractional rates, the token count is often stored as a float.
This allows smooth and continuous replenishment rather than discrete jumps.

### How I Implemented It

Each user has a bucket with:

  * A current token count.
  * The last refill timestamp.

On each request:

1.  Calculate the elapsed time since the last refill.
2.  Add tokens based on the elapsed time and the refill rate.
3.  Cap the tokens at the maximum capacity.
4.  If enough tokens are available, decrement and allow the request.
5.  Otherwise, reject it.

**Redis implementation:**

  * Similar logic, but implemented using a Lua script for atomic operations.
  * Steps:
    1.  Retrieve the bucket state (tokens and last refill time).
    2.  Calculate new tokens = `min(capacity, tokens + elapsed * refill_rate)`.
    3.  If enough tokens are available, decrement and update the values atomically (CAS).

### Edge Cases I Discovered

1.  **Clock skew:**

      * Problem: Inconsistent system clocks across servers can affect token refill accuracy.
      * Solution: Use Redis server time for all calculations.

2.  **Float precision:**

      * Problem: Floating-point arithmetic may cause precision issues (e.g., `9.999999` not equal to `10.0`).
      * Solution: Round token values before comparison.

### When to Use

  * Best for scenarios where **short bursts** of traffic should be allowed.
  * Ideal for **API rate limiting** and **distributed systems**.
  * Provides smoother flow control compared to the Fixed Window approach.

### Performance Characteristics

  * Constant time (**O(1)**) per operation.
  * **In-Memory (Storage Benchmark):** 217.4 ns/op (160 B/op, 4 allocs/op).
  * **Redis (Storage Benchmark):** 31360 ns/op (2608 B/op, 25 allocs/op).
  * **In-Memory (Multiple Keys, Concurrent):** 70.60 ns/op (160 B/op, 4 allocs/op).
  * Handles burst traffic gracefully.
  * Redis Lua scripts ensure atomic and consistent updates across nodes.

-----

## Sliding Window Algorithm

### Mental Model

I used the **Sliding Window Counter** version of this algorithm.
Unlike the Fixed Window, which only considers the current window, this method also factors in the previous window to provide smoother transitions.

For example, assume the limit is 10 requests per 10 seconds:

  * The previous window consumed 8 requests.
  * The current window (2 seconds into it) has consumed 2 requests.
  * The weight of the current window is `2/10 = 0.2`.

So the total weighted count is:
`(1 - 0.2) * previous + current = 0.8 * 8 + 2 = 8.4`.

If `(weighted_count + increment) > limit`, the request is rejected.

### How I Implemented It

1.  Retrieve the counts for both the current and previous windows.
2.  Calculate how much time has passed in the current window.
3.  Compute the weighted total:
    ```
    weighted = (1 - ratio) * previous + current
    ```
4.  If the weighted total plus the incoming request exceeds the limit, reject it.
5.  Otherwise, allow the request.

**Redis implementation:**

  * The same logic is implemented in a Lua script to ensure atomic reads and writes.

### Edge Cases I Discovered

1.  **Synchronization lag:**

      * Problem: Slight timing differences between window updates can cause double counting.
      * Solution: Use high-precision timestamps (milliseconds) to define window boundaries.

2.  **Temporary memory increase:**

      * Problem: Both current and previous windows are stored, increasing memory usage.
      * Solution: Set expiry for the previous window after one full window duration.

### When to Use

  * Ideal for **smooth throttling** and **distributed rate limiting**.
  * Provides better accuracy than Fixed Window.
  * Recommended when avoiding sudden traffic spikes is important.

### Performance Characteristics

  * **In-Memory (Storage Benchmark):** 336.1 ns/op (96 B/op, 6 allocs/op).
  * **Redis (Storage Benchmark):** 26817 ns/op (1624 B/op, 31 allocs/op).
  * **In-Memory (Single Key, Concurrent):** 48.50 ns/op (80 B/op, 6 allocs/op).
  * **In-Memory (Multiple Keys, Concurrent):** 63.62 ns/op (97 B/op, 6 allocs/op).
  * Slightly higher CPU usage due to dual window reads.
  * Still **O(1)** overall.
  * Balances accuracy, fairness, and performance effectively.