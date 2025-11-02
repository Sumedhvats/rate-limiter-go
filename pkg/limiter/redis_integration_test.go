package limiter

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"

	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func RedisTest(t *testing.T) (*storage.RedisMemory, func()) {
	ctx := context.Background()
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	assert.NoError(t, err)

	redisEndpint, err := redisContainer.Endpoint(ctx, "")
	assert.NoError(t, err)

	store := storage.NewRedisStorage(redisEndpint)

	cleanup := func() {
		redisContainer.Terminate(ctx)
	}
	return store, cleanup
}

func TestSlidingWindow_Redis_Concurrent(t *testing.T) {
	store, cleanup := RedisTest(t)
	defer cleanup()

	limiter := NewSlidingWindowLimiter(store, Config{
		Rate:   100,
		Window: time.Second,
	})
	var wg sync.WaitGroup
	var allowed, denied int64
	for i := 0; i < 10; i++ {

		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				ok, _ := limiter.Allow("test1")
				if ok {
					atomic.AddInt64(&allowed, 1)
				} else {
					atomic.AddInt64(&denied, 1)
				}
			}

		}()
	}
	wg.Wait()
	tolerance := 5.0

	assert.InDelta(t, int64(100), allowed, tolerance, "allowed count should be around 100")

	assert.InDelta(t, int64(100), denied, tolerance, "denied count should be around 100")

	assert.Equal(t, int64(200), allowed+denied, "total should be exactly 200")

}

func TestRedis_DestributedRateLimiting(t *testing.T) {
	store, cleanup := RedisTest(t)
	defer cleanup()

	limiter := NewSlidingWindowLimiter(store, Config{
		Rate:   10,
		Window: time.Minute,
	})
	limiter2 := NewSlidingWindowLimiter(store, Config{
		Rate:   10,
		Window: 1 * time.Minute,
	})

	for i := 0; i < 5; i++ {
		allowed, _ := limiter.Allow("user:123")
		assert.True(t, allowed)
	}
	for i := 0; i < 5; i++ {
		allowed, _ := limiter2.Allow("user:123")
		assert.True(t, allowed)
	}

	allowed, _ := limiter.Allow("user:123")
	assert.False(t, allowed)

	allowed, _ = limiter2.Allow("user:123")
	assert.False(t, allowed)
}
