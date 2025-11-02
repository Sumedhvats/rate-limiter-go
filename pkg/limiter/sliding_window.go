package limiter

import (
	"fmt"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
	"math"
	"time"
)

type SlidingWindowLimiter struct {
	storage Storage
	config  Config
}

func NewSlidingWindowLimiter(store Storage, cfg Config) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		storage: store,
		config:  cfg,
	}
}
func (swl *SlidingWindowLimiter) AllowN(key string, n int) (bool, error) {
	now := time.Now()
	windowStart := now.Truncate(swl.config.Window)
	currWinKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	prevStart := windowStart.Add(-swl.config.Window)
	prevWinKey := fmt.Sprintf("%s:%d", key, prevStart.Unix())

	if redisStore, ok := swl.storage.(*storage.RedisMemory); ok {
		elapsed := now.Sub(windowStart)
		weight := 1.0 - (float64(elapsed) / float64(swl.config.Window))
		return redisStore.SlidingWindowIncrement(
			currWinKey,
			prevWinKey,
			int(swl.config.Rate),
			weight,
			swl.config.Window*2,
		)
	}

	return swl.allowNMemory(currWinKey, prevWinKey, now, windowStart, n)
}

func (swl *SlidingWindowLimiter) allowNMemory(
	currWinKey, prevWinKey string,
	now, windowStart time.Time,
	n int,
) (bool, error) {
	current_data, _ := swl.storage.Get(currWinKey)
	previous_data, _ := swl.storage.Get(prevWinKey)

	var currCount, prevCount int64
	if current_data != nil {
		currCount = current_data.(int64)
	}
	if previous_data != nil {
		prevCount = previous_data.(int64)
	}

	currElapsedTime := now.Sub(windowStart)
	weight := 1 - (float64(currElapsedTime) / float64(swl.config.Window))
	if weight < 0 {
		weight = 0
	}
	if weight > 1 {
		weight = 1
	}

	weightedCount := math.Ceil((float64(prevCount) * weight) + float64(currCount))

	if weightedCount+float64(n) > float64(swl.config.Rate) {
		return false, nil
	}

	swl.storage.Increment(currWinKey, n, swl.config.Window*2)
	return true, nil
}

func (swl *SlidingWindowLimiter) Allow(key string) (bool, error) {
	return swl.AllowN(key, 1)
}

func (swl *SlidingWindowLimiter) Reset(key string) error {
	now := time.Now()
	windowStart := now.Truncate(swl.config.Window)
	prevStart := windowStart.Add(-swl.config.Window)

	currWinKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())
	prevWinKey := fmt.Sprintf("%s:%d", key, prevStart.Unix())

	_ = swl.storage.Delete(currWinKey)
	_ = swl.storage.Delete(prevWinKey)
	return nil
}

func (swl *SlidingWindowLimiter) GetStats(key string) (*stats, error) {
	now := time.Now()
	windowStart := now.Truncate(swl.config.Window)
	prevStart := windowStart.Add(-swl.config.Window)

	currWinKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())
	prevWinKey := fmt.Sprintf("%s:%d", key, prevStart.Unix())

	currData, _ := swl.storage.Get(currWinKey)
	prevData, _ := swl.storage.Get(prevWinKey)

	var currCount, prevCount int64
	if currData != nil {
		currCount = currData.(int64)
	}
	if prevData != nil {
		prevCount = prevData.(int64)
	}

	elapsed := now.Sub(windowStart)
	weight := 1 - (float64(elapsed) / float64(swl.config.Window))
	if weight < 0 {
		weight = 0
	}
	if weight > 1 {
		weight = 1
	}

	weightedCount := math.Ceil((float64(prevCount) * weight) + float64(currCount))
	remaining := int64(swl.config.Rate) - int64(weightedCount)
	if remaining < 0 {
		remaining = 0
	}

	resetAt := windowStart.Add(swl.config.Window)
	return &stats{
		Limit:     int(swl.config.Rate),
		Remaining: int(remaining),
		ResetAt:   resetAt,
	}, nil
}
