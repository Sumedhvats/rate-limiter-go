package limiter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func TestBasicRateLimiting(t *testing.T) {
	store := storage.NewMemoryStorage()
	config := limiter.Config{Rate: 10, Window: 1 * time.Minute, Burst: 10}
	limiter := limiter.NewTokenBucketLimiter(store, config)
	for i := 0; i < 10; i++ {
		ok, err := limiter.Allow("test1")
		assert.NoError(t, err)
		assert.True(t, ok)
	}
	ok, err := limiter.Allow("user1")
	assert.NoError(t, err)
	assert.True(t, ok)
}
func TestTokenRefil(t *testing.T) {
	store := storage.NewMemoryStorage()
	config := limiter.Config{Rate: 10, Window: 1 * time.Minute, Burst: 10}
	limiter := limiter.NewTokenBucketLimiter(store, config)
	for i := 0; i < 10; i++ {
		ok, err := limiter.Allow("test2")
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	time.Sleep(6 * time.Second)

	stats, err := limiter.GetStats("test2")
	assert.NoError(t, err)
	assert.Greater(t, stats.Remaining, 0)

	ok, _ := limiter.Allow("test2")
	assert.True(t, ok)
	ok, _ = limiter.Allow("test2")
	assert.False(t, ok)
}
func TestBurstHandling(t *testing.T) {
	store := storage.NewMemoryStorage()
	config := limiter.Config{Rate: 100, Window: time.Minute, Burst: 20}
	limiter := limiter.NewTokenBucketLimiter(store, config)
	for i := 0; i < 20; i++ {
		ok, err := limiter.Allow("test3")
		assert.NoError(t, err)
		assert.True(t, ok)
	}
	ok, _ := limiter.Allow("test3")
	assert.False(t, ok)
	time.Sleep(12 * time.Second)
	for i := 0; i < 20; i++ {
		ok, err := limiter.Allow("test3")
		assert.NoError(t, err)
		assert.True(t, ok)
	}
}
func TestAllowNLimiting(t *testing.T) {
	store := storage.NewMemoryStorage()
	config := limiter.Config{Rate: 100, Window: 1 * time.Minute, Burst: 100}
	limiter := limiter.NewTokenBucketLimiter(store, config)
	ok, err := limiter.AllowN("test4", 50)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = limiter.AllowN("test4", 50)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = limiter.AllowN("test4", 1)
	assert.NoError(t, err)
	assert.False(t, ok)

}
func TestMultipleKeys(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 10, Window: time.Minute, Burst: 10}
	limiter := limiter.NewTokenBucketLimiter(store, cfg)

	for i := 0; i < 10; i++ {
		limiter.Allow("user1")
	}

	ok, _ := limiter.Allow("user2")
	assert.True(t, ok)
}
