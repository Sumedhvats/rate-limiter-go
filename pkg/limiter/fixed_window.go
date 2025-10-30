package limiter

import (
	"fmt"
	"time"
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
func (f *FixedWindowLimiter) AllowN(key string, n int) (bool, error) {
	now := time.Now()
	windowStart := now.Truncate(f.config.Window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	count, err := f.storage.Increment(windowKey, n,f.config.Window)
	if err != nil {
		return false, err
	}
	if count == int64(n) {
		f.storage.Set(windowKey, count, f.config.Window)
	}
	return count <= int64(f.config.Rate), nil
}

func (f *FixedWindowLimiter) Allow(key string) (bool, error) {
	return f.AllowN(key, 1)
}
func (f *FixedWindowLimiter)Reset(key string) error{
	return f.storage.Delete(key)
}

func (f *FixedWindowLimiter) GetStats(key string)(*stats,error){
	now:=time.Now()
	windowStart := now.Truncate(f.config.Window)
	data , _ := f.storage.Get(key)
	if data == nil{
			return &stats{}, fmt.Errorf("not found key")
	}
	count:=data.(int64)
	resetAt := windowStart.Add(f.config.Window)
	return &stats{
	Limit:     f.config.Burst,
		Remaining: int(int64(f.config.Burst) - count),
		ResetAt:   resetAt,
	}, nil
}