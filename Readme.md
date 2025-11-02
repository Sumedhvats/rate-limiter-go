# rate-limiter-go

A production-ready, flexible rate limiting library for Go with multiple algorithms and storage backends.

[![Go Reference](https://pkg.go.dev/badge/github.com/sumedhvats/rate-limiter-go.svg)](https://pkg.go.dev/github.com/sumedhvats/rate-limiter-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sumedhvats/rate-limiter-go)](https://goreportcard.com/report/github.com/sumedhvats/rate-limiter-go)
[![Coverage](https://codecov.io/gh/sumedhvats/rate-limiter-go/branch/main/graph/badge.svg)](https://codecov.io/gh/sumedhvats/rate-limiter-go)

---

## Features

* **Multiple Rate Limiting Algorithms**

  * Token Bucket (smooth traffic shaping)
  * Fixed Window Counter (simple and fast)
  * Sliding Window Log (precise but memory-intensive)
  * Sliding Window Counter (balanced and recommended)

* **Flexible Storage Backends**

  * In-memory (for single-instance use)
  * Redis (for distributed systems)
  * Extensible interface for custom backends

* **High Performance**

  * Atomic operations with Lua scripts (Redis)
  * Concurrent-safe with minimal locking
  * Benchmarked at high throughput (details below)

* **Production Ready**

  * Comprehensive test coverage (>90%)
  * HTTP middleware support
  * Proven algorithmic reliability

---

## Installation

```bash
go get github.com/sumedhvats/rate-limiter-go
```

---

## Quick Start

### Basic Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func main() {
	store := storage.NewMemoryStorage()

	rateLimiter := limiter.NewSlidingWindowLimiter(store, limiter.Config{
		Rate:   10,
		Window: 1 * time.Minute,
	})

	allowed, err := rateLimiter.Allow("user:123")
	if err != nil {
		panic(err)
	}

	if !allowed {
		fmt.Println("Rate limit exceeded!")
		return
	}

	fmt.Println("Request allowed!")
}
```

---

### HTTP Middleware

```go
package main

import (
	"net/http"
	"time"

	"github.com/sumedhvats/rate-limiter-go/middleware"
	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func main() {
	store := storage.NewRedisStorage("localhost:6379")

	rateLimiter := limiter.NewSlidingWindowLimiter(store, limiter.Config{
		Rate:   100,
		Window: 1 * time.Minute,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Data served successfully"))
	})

	handler := middleware.RateLimitMiddleware(middleware.Config{
		Limiter: rateLimiter,
	})(mux)

	http.ListenAndServe(":8080", handler)
}
```

---

## Algorithm Comparison

| Algorithm              | Accuracy  | Memory Usage | Recommended Use Case                       |
| ---------------------- | --------- | ------------ | ------------------------------------------ |
| Token Bucket           | High      | Low          | API gateways, smooth traffic shaping       |
| Fixed Window Counter   | Medium    | Very Low     | Simple counting, analytics                 |
| Sliding Window Log     | Very High | High         | Financial systems, strict compliance cases |
| Sliding Window Counter | High      | Low          | General-purpose, balanced rate limiting    |

> Note: Fixed Window has boundary issues (bursts near window edges).

---

## Advanced Usage

### Per-User Rate Limiting

```go
func apiHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r) // from JWT, cookie, etc.

	allowed, err := rateLimiter.Allow(fmt.Sprintf("user:%s", userID))
	if err != nil {
		http.Error(w, "Internal error", 500)
		return
	}

	if !allowed {
		http.Error(w, "Rate limit exceeded", 429)
		return
	}

	// Handle request
}
```

---

### Different Limits per User Tier

```go
premiumLimiter := limiter.NewSlidingWindowLimiter(store, limiter.Config{
	Rate:   1000,
	Window: 1 * time.Minute,
})

freeLimiter := limiter.NewSlidingWindowLimiter(store, limiter.Config{
	Rate:   100,
	Window: 1 * time.Minute,
})

func selectLimiter(user User) limiter.Limiter {
	if user.IsPremium {
		return premiumLimiter
	}
	return freeLimiter
}
```

---

### Token Bucket with Burst Handling

```go
rateLimiter := limiter.NewTokenBucketLimiter(store, limiter.Config{
	Rate:   100,
	Window: 1 * time.Minute,
	Burst:  50, // allow bursts of up to 50 requests
})
```

---

### Distributed Rate Limiting with Redis

```go
store := storage.NewRedisStorage("redis-cluster:6379")

rateLimiter := limiter.NewSlidingWindowLimiter(store, limiter.Config{
	Rate:   10000,
	Window: 1 * time.Minute,
})

// shared across multiple app instances
```

---

## Testing

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./benchmarks
```

---

## Benchmarks

```
BenchmarkTokenBucket_Sequential-8        500000    2340 ns/op    384 B/op    8 allocs/op
BenchmarkSlidingWindow_Sequential-8      400000    2890 ns/op    512 B/op   10 allocs/op
BenchmarkTokenBucket_Concurrent-8       2000000     890 ns/op    128 B/op    4 allocs/op
```

---

## Project Structure

```
rate-limiter-go/
├── pkg/
│   ├── limiter/          # Algorithms
│   │   ├── limiter.go
│   │   ├── token_bucket.go
│   │   ├── fixed_window.go
│   │   └── sliding_window.go
│   └── storage/          # Storage backends
│       ├── storage.go
│       ├── memory.go
│       └── redis.go
├── middleware/           # HTTP middleware
├── examples/             # Usage examples
└── benchmarks/           # Performance tests
```

---

## Contributing

Contributions are welcome. To contribute:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/feature-name`)
3. Write tests for your changes
4. Ensure tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add feature-name'`)
6. Push to your fork and open a Pull Request

---

## License

MIT License – see [LICENSE](LICENSE) for details.

---

## Acknowledgments

* Inspired by Cloudflare’s rate limiting architecture
* Lua scripting patterns adapted from [redis.io documentation](https://redis.io/docs/manual/programmability/)

