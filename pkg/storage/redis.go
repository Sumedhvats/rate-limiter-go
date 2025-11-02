package storage

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMemory struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStorage(addr string) *RedisMemory {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second})

	if err := client.Ping(ctx).Err(); err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	return &RedisMemory{
		client: client,
		ctx:    ctx,
	}
}
func (r *RedisMemory) Get(key string) (interface{}, error) {
	data, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if num, convErr := strconv.ParseInt(data, 10, 64); convErr == nil {
		return num, nil
	}
	return data, nil
}
func (r *RedisMemory) Set(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisMemory) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisMemory) Increment(key string, amount int, ttl time.Duration) (int64, error) {
	return r.client.IncrBy(r.ctx, key, int64(amount)).Result()
}
func (r *RedisMemory) SlidingWindowIncrement(
	currentKey, previousKey string,
	limit int,
	weight float64,
	ttl time.Duration,

) (bool, error) {
	script := redis.NewScript(`
-- KEYS[1]: current window key
-- KEYS[2]: previous window key
-- ARGV[1]: limit
-- ARGV[2]: weight
-- ARGV[3]: TTL in seconds
-- ARGV[4]: increment

local current_key = KEYS[1]
local previous_key = KEYS[2]
local limit = tonumber(ARGV[1])
local weight = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])
local increment = tonumber(ARGV[4])

local current = tonumber(redis.call('GET', current_key) or '0')
local previous = tonumber(redis.call('GET', previous_key) or '0')

local weighted_count = math.floor(previous * weight) + current
if weighted_count + increment > limit then
    return 0
end

redis.call('INCRBY', current_key, increment)
redis.call('EXPIRE', current_key, ttl)
return 1

`)
	result, err := script.Run(
		r.ctx,
		r.client,
		[]string{currentKey, previousKey},
		limit,
		weight,
		int(ttl.Seconds()),
		1,
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, err
}
func (r *RedisMemory) FixedWindowIncrement(
	key string,
	increment int,
	limit int,
	ttl int,
) (bool, error) {
	script := redis.NewScript(`
-- Fixed Window Rate Limiter
-- KEYS[1]: window key
-- ARGV[1]: increment amount
-- ARGV[2]: rate limit
-- ARGV[3]: TTL in seconds

local key = KEYS[1]
local increment = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

-- Get current count
local current = tonumber(redis.call('GET', key) or '0')

-- Check if incrementing would exceed limit
if current + increment > limit then
    return 0  -- Denied
end

-- Increment and set expiry
redis.call('INCRBY', key, increment)
redis.call('EXPIRE', key, ttl)

return 1  -- Allowed
`)

	result, err := script.Run(
		r.ctx,
		r.client,
		[]string{key},
		increment,
		limit,
		ttl,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
func (r *RedisMemory) TokenBucketAllow(
	key string,
	tokens int,
	capacity int,
	refillRate float64,
	nowUnix int64,
	ttl int,
) (bool, error) {
	script := redis.NewScript(`
-- Token Bucket Rate Limiter
-- KEYS[1]: bucket key
-- ARGV[1]: tokens to consume
-- ARGV[2]: bucket capacity
-- ARGV[3]: refill rate (tokens per second)
-- ARGV[4]: current timestamp (unix)
-- ARGV[5]: TTL in seconds

local key = KEYS[1]
local tokens_to_consume = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill_rate = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

-- Get current bucket state
-- Format: "tokens:last_refill_time"
local bucket = redis.call('GET', key)

local current_tokens
local last_refill

if bucket == false then
    -- New bucket - start with full capacity
    current_tokens = capacity
    last_refill = now
else
    -- Parse existing bucket
    local colon_pos = string.find(bucket, ":")
    current_tokens = tonumber(string.sub(bucket, 1, colon_pos - 1))
    last_refill = tonumber(string.sub(bucket, colon_pos + 1))
end

-- Calculate tokens to add based on time elapsed
local elapsed = now - last_refill
local tokens_to_add = elapsed * refill_rate

-- Refill tokens (capped at capacity)
current_tokens = math.min(current_tokens + tokens_to_add, capacity)

-- Check if we have enough tokens
if current_tokens >= tokens_to_consume then
    -- Consume tokens
    current_tokens = current_tokens - tokens_to_consume
    
    -- Save new state
    local new_bucket = string.format("%.6f:%d", current_tokens, now)
    redis.call('SETEX', key, ttl, new_bucket)
    
    return 1  -- Allowed
else
    -- Not enough tokens - update last refill time anyway
    local new_bucket = string.format("%.6f:%d", current_tokens, now)
    redis.call('SETEX', key, ttl, new_bucket)
    
    return 0  -- Denied
end
`)

	result, err := script.Run(
		r.ctx,
		r.client,
		[]string{key},
		tokens,
		capacity,
		refillRate,
		nowUnix,
		ttl,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
