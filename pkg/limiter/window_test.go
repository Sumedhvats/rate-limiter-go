package limiter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func TestBasicFixedWindow(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 10, Window: time.Minute, Burst: 10}
	fixed := limiter.NewFixedWindowLimiter(store, cfg)

	for i := 0; i < 10; i++ {
		ok, err := fixed.Allow("user1")
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	ok, err := fixed.Allow("user1")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestWindowReset(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 5, Window: 2 * time.Second, Burst: 5}
	fixed := limiter.NewFixedWindowLimiter(store, cfg)

	for i := 0; i < 5; i++ {
		ok, _ := fixed.Allow("user2")
		assert.True(t, ok)
	}

	ok, _ := fixed.Allow("user2")
	assert.False(t, ok)

	time.Sleep(3 * time.Second)

	ok, _ = fixed.Allow("user2")
	assert.True(t, ok, "Should reset after window expires")
}

func TestFixedWindowBoundary(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 10, Window: 5 * time.Second, Burst: 10}
	fixed := limiter.NewFixedWindowLimiter(store, cfg)

	for i := 0; i < 10; i++ {
		ok, err := fixed.Allow("boundaryUser")
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	time.Sleep(6 * time.Second)

	successCount := 0
	for i := 0; i < 10; i++ {
		ok, _ := fixed.Allow("boundaryUser")
		if ok {
			successCount++
		}
	}

	assert.LessOrEqual(t, successCount, 10, "Fixed Window allows bursts across boundaries")
}

func TestAllowNFixedWindow(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 10, Window: 10 * time.Second, Burst: 10}
	fixed := limiter.NewFixedWindowLimiter(store, cfg)

	ok, err := fixed.AllowN("batchUser", 5)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = fixed.AllowN("batchUser", 5)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = fixed.AllowN("batchUser", 1)
	assert.NoError(t, err)
	assert.False(t, ok, "Should reject since burst exhausted")
}

func TestMultipleUsersFixedWindow(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 5, Window: 5 * time.Second, Burst: 5}
	fixed := limiter.NewFixedWindowLimiter(store, cfg)

	for i := 0; i < 5; i++ {
		ok, _ := fixed.Allow("userA")
		assert.True(t, ok)
	}

	ok, _ := fixed.Allow("userA")
	assert.False(t, ok)

	ok, _ = fixed.Allow("userB") // new user should be allowed
	assert.True(t, ok)
}

func TestSmallWindowBehavior(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 3, Window: 1 * time.Second}
	fixed := limiter.NewFixedWindowLimiter(store, cfg)

	for i := 0; i < 3; i++ {
		ok, _ := fixed.Allow("tinyUser")
		assert.True(t, ok)
	}

	ok, _ := fixed.Allow("tinyUser")
	assert.False(t, ok)

	time.Sleep(1100 * time.Millisecond) // wait for new window
	ok, _ = fixed.Allow("tinyUser")
	assert.True(t, ok)
}
func TestBasicSlidingWindow(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 10, Window: time.Minute}
	fixed := limiter.NewSlidingWindowLimiter(store, cfg)

	for i := 0; i < 10; i++ {
		ok, err := fixed.Allow("user1")
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	ok, err := fixed.Allow("user1")
	assert.NoError(t, err)
	assert.False(t, ok)
}
func TestSlidingWindowReset(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := limiter.Config{Rate: 5, Window: 2 * time.Second, Burst: 5}
	fixed := limiter.NewSlidingWindowLimiter(store, cfg)

	for i := 0; i < 5; i++ {
		ok, _ := fixed.Allow("user2")
		assert.True(t, ok)
	}

	ok, _ := fixed.Allow("user2")
	assert.False(t, ok)

	time.Sleep(3 * time.Second)

	ok, _ = fixed.Allow("user2")
	assert.True(t, ok, "Should reset after window expires")
}

func TestSlidingWindowBurstHandling(t *testing.T) {
    store := storage.NewMemoryStorage()
    config := limiter.Config{Rate: 20, Window: 10*time.Second}
    limiter := limiter.NewSlidingWindowLimiter(store, config)
    
    for i := 0; i < 20; i++ {
        ok, err := limiter.Allow("test3")
        assert.NoError(t, err)
        assert.True(t, ok)
    }
    ok, _ := limiter.Allow("test3")
    assert.False(t, ok)
    
    // Wait for MORE than a full window to completely reset
    time.Sleep(21 * time.Second)
    
    for i := 0; i < 20; i++ {
        ok, err := limiter.Allow("test3")
        assert.NoError(t, err)
        assert.True(t, ok)
    }
}

func TestSlidingWIndowAllowNLimiting(t *testing.T) {
	store := storage.NewMemoryStorage()
	config := limiter.Config{Rate: 100, Window: 1 * time.Minute, Burst: 100}
	limiter := limiter.NewSlidingWindowLimiter(store, config)
	ok, err :=limiter.AllowN("test4",50)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err =limiter.AllowN("test4",50)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err =limiter.AllowN("test4",1)
	assert.NoError(t, err)
	assert.False(t, ok)

}


func TestSWMultipleKeys(t *testing.T) {
store := storage.NewMemoryStorage()
cfg := limiter.Config{Rate: 10, Window: 5*time.Second, Burst: 10}
limiter := limiter.NewSlidingWindowLimiter(store, cfg)

for i := 0; i < 10; i++ {
	limiter.Allow("user1")
}

ok, _ := limiter.Allow("user2")
assert.True(t, ok)
}