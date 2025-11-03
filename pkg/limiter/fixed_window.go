package limiter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

type FixedWindowLimiter struct {
	storage Storage
	config  Config
}

func NewFixedWindowLimiter(store Storage, cfg Config) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		storage: store,
		config:  cfg,
	}
}
func (fwl *FixedWindowLimiter) AllowN(key string, n int) (bool, error) {
	// now := time.Now()
	// windowStart := now.Truncate(fwl.config.Window)
	// windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	nowUnix := time.Now().Unix()
	//(nowUnix / windowSize) * windowSize
	windowStart := (nowUnix / int64(fwl.config.Window.Seconds())) * int64(fwl.config.Window.Seconds())
	
	// Pre-allocate key buffer or use string builder


	windowKey := key + ":" + strconv.FormatInt(windowStart, 10)

	if redisStore, ok := fwl.storage.(*storage.RedisMemory); ok {
		return fwl.allowNRedis(redisStore, windowKey, n)
	}

	return fwl.allowNMemory(windowKey, n)
}

func (fwl *FixedWindowLimiter) allowNRedis(store *storage.RedisMemory, windowKey string, n int) (bool, error) {
	return store.FixedWindowIncrement(
		windowKey,
		n,
		int(fwl.config.Rate),
		int(fwl.config.Window.Seconds())*2,
	)
}

func (fwl *FixedWindowLimiter) allowNMemory(windowKey string, n int) (bool, error) {
	newCount, err := fwl.storage.Increment(windowKey, n, fwl.config.Window*2)
	if err != nil {
		return false, err
	}
	if newCount <= int64(fwl.config.Rate) {
		return true, nil
	}

	fwl.storage.Increment(windowKey, -n, fwl.config.Window*2)
	return false, nil
}

func (f *FixedWindowLimiter) Allow(key string) (bool, error) {
	return f.AllowN(key, 1)
}
func (f *FixedWindowLimiter) Reset(key string) error {
	return f.storage.Delete(key)
}

func (f *FixedWindowLimiter) GetStats(key string) (*stats, error) {
	now := time.Now()
	windowStart := now.Truncate(f.config.Window)
	data, _ := f.storage.Get(key)
	if data == nil {
		return &stats{}, fmt.Errorf("not found key")
	}
	count := data.(int64)
	resetAt := windowStart.Add(f.config.Window)
	return &stats{
		Limit:     f.config.Burst,
		Remaining: int(int64(f.config.Burst) - count),
		ResetAt:   resetAt,
	}, nil
}
