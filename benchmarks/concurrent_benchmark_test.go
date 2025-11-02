package benchmarks

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("user-%d", i)
	}
	return keys
}

func BenchmarkSingleKey(b *testing.B) {
	cfg := limiter.Config{
		Rate:   100,
		Window: time.Second,
	}

	b.Run("TokenBucket/Sequential", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Allow("user1")
		}
	})

	b.Run("TokenBucket/Concurrent", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Allow("user1")
			}
		})
	})

	b.Run("FixedWindow/Sequential", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Allow("user1")
		}
	})

	b.Run("FixedWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Allow("user1")
			}
		})
	})

	b.Run("SlidingWindow/Sequential", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Allow("user1")
		}
	})

	b.Run("SlidingWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Allow("user1")
			}
		})
	})
}

func BenchmarkMultipleKeys(b *testing.B) {
	cfg := limiter.Config{
		Rate:   100,
		Window: time.Second,
	}
	const keyPoolSize = 1000
	keys := generateKeys(keyPoolSize)

	b.Run("TokenBucket/Sequential", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[i%keyPoolSize]
			l.Allow(key)
		}
	})

	b.Run("TokenBucket/Concurrent", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		var counter uint64
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := atomic.AddUint64(&counter, 1)
				key := keys[idx%keyPoolSize]
				l.Allow(key)
			}
		})
	})

	b.Run("FixedWindow/Sequential", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[i%keyPoolSize]
			l.Allow(key)
		}
	})

	b.Run("FixedWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		var counter uint64
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := atomic.AddUint64(&counter, 1)
				key := keys[idx%keyPoolSize]
				l.Allow(key)
			}
		})
	})

	b.Run("SlidingWindow/Sequential", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[i%keyPoolSize]
			l.Allow(key)
		}
	})

	b.Run("SlidingWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		var counter uint64
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := atomic.AddUint64(&counter, 1)
				key := keys[idx%keyPoolSize]
				l.Allow(key)
			}
		})
	})
}

func BenchmarkRandomDistribution(b *testing.B) {
	cfg := limiter.Config{
		Rate:   100,
		Window: time.Second,
	}
	const keyPoolSize = 1000
	keys := generateKeys(keyPoolSize)

	b.Run("TokenBucket/Sequential", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		rng := rand.New(rand.NewSource(42))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[rng.Intn(keyPoolSize)]
			l.Allow(key)
		}
	})

	b.Run("TokenBucket/Concurrent", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for pb.Next() {
				key := keys[rng.Intn(keyPoolSize)]
				l.Allow(key)
			}
		})
	})

	b.Run("FixedWindow/Sequential", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		rng := rand.New(rand.NewSource(42))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[rng.Intn(keyPoolSize)]
			l.Allow(key)
		}
	})

	b.Run("FixedWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for pb.Next() {
				key := keys[rng.Intn(keyPoolSize)]
				l.Allow(key)
			}
		})
	})

	b.Run("SlidingWindow/Sequential", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		rng := rand.New(rand.NewSource(42))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[rng.Intn(keyPoolSize)]
			l.Allow(key)
		}
	})

	b.Run("SlidingWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for pb.Next() {
				key := keys[rng.Intn(keyPoolSize)]
				l.Allow(key)
			}
		})
	})
}

func BenchmarkHotspotDistribution(b *testing.B) {
	cfg := limiter.Config{
		Rate:   100,
		Window: time.Second,
	}
	const keyPoolSize = 1000
	const hotspotSize = 200
	keys := generateKeys(keyPoolSize)

	hotspotKey := func(rng *rand.Rand) string {
		if rng.Intn(100) < 80 {
			return keys[rng.Intn(hotspotSize)]
		}
		return keys[hotspotSize+rng.Intn(keyPoolSize-hotspotSize)]
	}

	b.Run("TokenBucket/Sequential", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		rng := rand.New(rand.NewSource(42))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Allow(hotspotKey(rng))
		}
	})

	b.Run("TokenBucket/Concurrent", func(b *testing.B) {
		l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for pb.Next() {
				l.Allow(hotspotKey(rng))
			}
		})
	})

	b.Run("FixedWindow/Sequential", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		rng := rand.New(rand.NewSource(42))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Allow(hotspotKey(rng))
		}
	})

	b.Run("FixedWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for pb.Next() {
				l.Allow(hotspotKey(rng))
			}
		})
	})

	b.Run("SlidingWindow/Sequential", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		rng := rand.New(rand.NewSource(42))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Allow(hotspotKey(rng))
		}
	})

	b.Run("SlidingWindow/Concurrent", func(b *testing.B) {
		l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for pb.Next() {
				l.Allow(hotspotKey(rng))
			}
		})
	})
}

func BenchmarkScalability(b *testing.B) {
	cfg := limiter.Config{
		Rate:   100,
		Window: time.Second,
	}

	keySizes := []int{10, 100, 1000, 10000}

	for _, keySize := range keySizes {
		keys := generateKeys(keySize)

		b.Run(fmt.Sprintf("TokenBucket/Keys-%d", keySize), func(b *testing.B) {
			l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%uint64(keySize)]
					l.Allow(key)
				}
			})
		})

		b.Run(fmt.Sprintf("FixedWindow/Keys-%d", keySize), func(b *testing.B) {
			l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%uint64(keySize)]
					l.Allow(key)
				}
			})
		})

		b.Run(fmt.Sprintf("SlidingWindow/Keys-%d", keySize), func(b *testing.B) {
			l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%uint64(keySize)]
					l.Allow(key)
				}
			})
		})
	}
}

func BenchmarkDifferentRates(b *testing.B) {
	rates := []int{10, 100, 1000, 10000}
	const keyPoolSize = 100
	keys := generateKeys(keyPoolSize)

	for _, rate := range rates {
		cfg := limiter.Config{
			Rate:   rate,
			Window: time.Second,
		}

		b.Run(fmt.Sprintf("TokenBucket/Rate-%d", rate), func(b *testing.B) {
			l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%keyPoolSize]
					l.Allow(key)
				}
			})
		})

		b.Run(fmt.Sprintf("FixedWindow/Rate-%d", rate), func(b *testing.B) {
			l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%keyPoolSize]
					l.Allow(key)
				}
			})
		})

		b.Run(fmt.Sprintf("SlidingWindow/Rate-%d", rate), func(b *testing.B) {
			l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%keyPoolSize]
					l.Allow(key)
				}
			})
		})
	}
}

func BenchmarkVaryingConcurrency(b *testing.B) {
	cfg := limiter.Config{
		Rate:   100,
		Window: time.Second,
	}
	const keyPoolSize = 1000
	keys := generateKeys(keyPoolSize)

	concurrencyLevels := []int{1, 2, 4, 8, 16, 32}

	for _, procs := range concurrencyLevels {
		b.Run(fmt.Sprintf("TokenBucket/Procs-%d", procs), func(b *testing.B) {
			l := limiter.NewTokenBucketLimiter(storage.NewMemoryStorage(), cfg)
			b.SetParallelism(procs)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%keyPoolSize]
					l.Allow(key)
				}
			})
		})

		b.Run(fmt.Sprintf("FixedWindow/Procs-%d", procs), func(b *testing.B) {
			l := limiter.NewFixedWindowLimiter(storage.NewMemoryStorage(), cfg)
			b.SetParallelism(procs)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%keyPoolSize]
					l.Allow(key)
				}
			})
		})

		b.Run(fmt.Sprintf("SlidingWindow/Procs-%d", procs), func(b *testing.B) {
			l := limiter.NewSlidingWindowLimiter(storage.NewMemoryStorage(), cfg)
			b.SetParallelism(procs)
			var counter uint64
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := atomic.AddUint64(&counter, 1)
					key := keys[idx%keyPoolSize]
					l.Allow(key)
				}
			})
		})
	}
}
