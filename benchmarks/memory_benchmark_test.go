package benchmarks

import (
	"testing"
	"time"

	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func BenchmarkLimiterWithStorage(b *testing.B) {
	memStore := storage.NewMemoryStorage()
	redisStore := storage.NewRedisStorage("localhost:6379")

	tests := []struct {
		name  string
		store limiter.Storage
	}{
		{"Memory", memStore},
		{"Redis", redisStore},
	}

	for _, tt := range tests {
		b.Run(tt.name+"/TokenBucket", func(b *testing.B) {
			lim := limiter.NewTokenBucketLimiter(tt.store, limiter.Config{
				Rate:   100,
				Burst:  50,
				Window: time.Second,
			})
			key := "user123"

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				lim.Allow(key)
			}
		})

		b.Run(tt.name+"/FixedWindow", func(b *testing.B) {
			lim := limiter.NewFixedWindowLimiter(tt.store, limiter.Config{
				Rate:   100,
				Window: time.Second,
			})
			key := "user123"

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				lim.Allow(key)
			}
		})

		b.Run(tt.name+"/SlidingWindow", func(b *testing.B) {
			lim := limiter.NewSlidingWindowLimiter(tt.store, limiter.Config{
				Rate:   100,
				Window: time.Second,
			})
			key := "user123"

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				lim.Allow(key)
			}
		})
	}
}
