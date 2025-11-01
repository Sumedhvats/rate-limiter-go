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

func NewRedisMemory(addr string) *RedisMemory {
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
