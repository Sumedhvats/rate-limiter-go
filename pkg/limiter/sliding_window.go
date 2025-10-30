package limiter

import (
	"fmt"
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

	current_data, _ := swl.storage.Get(currWinKey)
	previous_data, _ := swl.storage.Get(prevWinKey)

	var currCount, prevCount int64
	if current_data != nil {
		currCount = current_data.(int64)
	}else{
		currCount=0
	}
	if previous_data != nil {
		prevCount = previous_data.(int64)
	}else{
		prevCount=0
	}

	currElapsedTime := now.Sub(windowStart)
	weight := 1 - (float64(currElapsedTime) / float64(swl.config.Window))

	weightedCount := math.Ceil((float64(prevCount) * weight) + float64(currCount))

	if weightedCount > float64(swl.config.Rate) {
		return false, nil
	}
	swl.storage.Increment(currWinKey, n,swl.config.Window)
	swl.storage.Set(currWinKey ,currCount+int64(n), swl.config.Window*2)
	return true, nil

}
