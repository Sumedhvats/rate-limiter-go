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
