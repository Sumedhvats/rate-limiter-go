// Package limiter provides rate limiting algorithm implementations.
package limiter

import "time"

// stats holds the current rate limit statistics for a key.
type stats struct {
	// Limit is the configured rate limit (e.g., Burst for token bucket).
	Limit     int
	// Remaining is the number of requests remaining in the current window.
	Remaining int
	// ResetAt is the time when the rate limit window resets.
	ResetAt   time.Time
}

// Limiter is the interface for a rate limiter.
type Limiter interface {
	// Allow checks if a single request (n=1) is allowed for the given key.
	Allow(key string) (bool, error)
	// AllowN checks if n requests are allowed for the given key.
	AllowN(key string, n int) (bool, error)
	// Reset clears the rate limit data for the given key.
	Reset(key string) error
	// GetStats returns the current rate limit statistics for the given key.
	GetStats(key string) (*stats, error)
}

// Config holds the configuration for a rate limiter.
type Config struct {
	// Rate is the number of requests allowed per window.
	Rate   int
	// Window is the time duration of the rate limit window.
	Window time.Duration
	// Burst is the maximum number of requests allowed in a burst.
	// This is typically used by Token Bucket algorithms.
	Burst  int
}

// Storage is the interface for storing rate limit data.
type Storage interface {
	// Get retrieves a value from the store by key.
	Get(key string) (interface{}, error)
	// Set stores a value in the store with a specified TTL.
	Set(key string, value interface{}, ttl time.Duration) error
	// Increment atomically increments a key's value and sets its TTL.
	Increment(key string, value int, ttl time.Duration) (int64, error)
	// Delete removes a key from the store.
	Delete(key string) error
}