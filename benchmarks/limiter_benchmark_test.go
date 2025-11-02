package benchmarks

import (
	"testing"
	"time"

	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)
func benchmarkLimiter(b *testing.B,l limiter.Limiter){
	key:="user123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Allow(key)
	}
}

func BenchmarkLimiters(b *testing.B) {
	memStore := storage.NewMemoryStorage()
	rate := 100
	window := 1 * time.Second

	b.Run("TokenBucket", func(b *testing.B) {
		tb := limiter.NewTokenBucketLimiter(memStore, limiter.Config{
			Rate:rate,
			Window: window,
		})
		benchmarkLimiter(b, tb)
	})

	b.Run("FixedWindow", func(b *testing.B) {
		fw := limiter.NewFixedWindowLimiter(memStore, limiter.Config{
			Rate:rate,
			Window: window,
		})
		benchmarkLimiter(b, fw)
	})

	b.Run("SlidingWindow", func(b *testing.B) {
		sw := limiter.NewSlidingWindowLimiter(memStore, limiter.Config{
			Rate:rate,
			Window: window,
		})
		benchmarkLimiter(b, sw)
	})
}