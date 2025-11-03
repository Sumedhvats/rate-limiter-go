// Package limiter provides rate limiting algorithm implementations.
package limiter

import (
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
	"time"
)

type tokenBucket struct {
	tokens        float64
	lastRefilTime time.Time
	capacity      int
	refillRate    float64
}
// TokenBucketLimiter implements the token bucket rate limiting algorithm.
type TokenBucketLimiter struct {
	storage Storage
	config  Config
}
// NewTokenBucketLimiter creates a new TokenBucketLimiter.
func NewTokenBucketLimiter(store Storage, cfg Config) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		storage: store,
		config:  cfg,
	}
}
// AllowN checks if n tokens can be consumed for the given key.
func (t *TokenBucketLimiter) AllowN(key string, n int) (bool, error) {
	if redisStore, ok := t.storage.(*storage.RedisMemory); ok {
		return t.allowNRedis(redisStore, key, n)
	}

	return t.allowNMemory(key, n)
}

func (t *TokenBucketLimiter) allowNRedis(store *storage.RedisMemory, key string, n int) (bool, error) {
	now := time.Now().Unix()
	refillRate := float64(t.config.Rate) / float64(t.config.Window.Seconds())

	return store.TokenBucketAllow(
		key,
		n,
		t.config.Burst,
		refillRate,
		now,
		int(t.config.Window.Seconds())*2,
	)
}

func (t *TokenBucketLimiter) allowNMemory(key string, n int) (bool, error) {
	now := time.Now()
	refillRate := float64(t.config.Rate) / float64(t.config.Window.Seconds())

	for {
		data, err := t.storage.Get(key)
		if err != nil {
			return false, err
		}

		var bucket *tokenBucket
		if data == nil {
			bucket = &tokenBucket{
				tokens:        float64(t.config.Burst),
				lastRefilTime: now,
				capacity:      t.config.Burst,
				refillRate:    refillRate,
			}
		} else {
			existingBucket := data.(*tokenBucket)
			bucket = &tokenBucket{
				tokens:        existingBucket.tokens,
				lastRefilTime: existingBucket.lastRefilTime,
				capacity:      existingBucket.capacity,
				refillRate:    existingBucket.refillRate,
			}

			elapsed := now.Sub(bucket.lastRefilTime).Seconds()
			tokenAdded := elapsed * bucket.refillRate
			bucket.tokens = min(bucket.tokens+tokenAdded, float64(bucket.capacity))
			bucket.lastRefilTime = now
		}

		if bucket.tokens >= float64(n) {
			bucket.tokens -= float64(n)

			if data == nil {
				err := t.storage.Set(key, bucket, t.config.Window*2)
				if err != nil {
					continue // Retry
				}
				return true, nil
			} else {
				t.storage.Set(key, bucket, t.config.Window*2)
				return true, nil
			}
		} else {
			t.storage.Set(key, bucket, t.config.Window*2)
			return false, nil
		}
	}
}
// Allow checks if a single request (1 token) can be consumed for the given key.
func (t *TokenBucketLimiter) Allow(key string) (bool, error) {
	return t.AllowN(key, 1)
}
// Reset clears the rate limit data for the given key.
func (t *TokenBucketLimiter) Reset(key string) error {
	return t.storage.Delete(key)
}
// GetStats returns the current rate limit statistics for the given key.
func (t *TokenBucketLimiter) GetStats(key string) (*stats, error) {
	now := time.Now()
	data, _ := t.storage.Get(key)

	if data == nil {
		return &stats{
			Limit:     t.config.Burst,
			Remaining: t.config.Burst,
			ResetAt:   now.Add(t.config.Window),
		}, nil
	}

	bucket := data.(*tokenBucket)
	elapsed := now.Sub(bucket.lastRefilTime).Seconds()
	tokenAdded := elapsed * bucket.refillRate
	currentTokens := min(bucket.tokens+tokenAdded, float64(bucket.capacity))

	return &stats{
		Limit:     t.config.Burst,
		Remaining: int(currentTokens),
		ResetAt:   bucket.lastRefilTime.Add(t.config.Window),
	}, nil
}
