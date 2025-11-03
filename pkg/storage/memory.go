// Package storage provides interfaces and implementations 
// for rate limiter data storage.
package storage

import (
	"errors"

	"sync"
	"time"
)
// MemoryStorage implements a storage backend using a thread-safe in-memory map.
type MemoryStorage struct {
	data    sync.Map
	cleanup *time.Ticker
}

type memoryEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryStorage creates and returns a new MemoryStorage.
// It also starts a background goroutine to clean up expired entries.
func NewMemoryStorage() *MemoryStorage {
	s := &MemoryStorage{
		cleanup: time.NewTicker(1 * time.Minute),
	}
	go s.cleanupExpired()
	return s
}
func (s *MemoryStorage) cleanupExpired() {
	for range s.cleanup.C {
		now := time.Now()
		s.data.Range(func(key, value interface{}) bool {
			entry := value.(*memoryEntry)
			if now.After(entry.expiresAt) {
				s.data.Delete(key)
			}
			return true
		})

	}
}

// Get retrieves a value from the in-memory store by key.
func (s *MemoryStorage) Get(key string) (interface{}, error) {
	val, ok := s.data.Load(key)
	if !ok {
		return nil, nil
	}
	entry := val.(*memoryEntry)
	if time.Now().After(entry.expiresAt) {
		s.data.Delete(key)
		return nil, nil
	}
	return entry.value, nil
}
// Set stores a value in the in-memory store with a specified TTL.
func (s *MemoryStorage) Set(key string, value interface{}, ttl time.Duration) error {
	now := time.Now()
	entry := &memoryEntry{
		value:     value,
		expiresAt: now.Add(ttl),
	}
	s.data.Store(key, entry)
	return nil
}
// Delete removes a key from the in-memory store.
func (s *MemoryStorage) Delete(key string) error {
	s.data.Delete(key)
	return nil
}
// Increment atomically increments a key's value in the in-memory store.
// It uses a Compare-And-Swap loop to handle concurrency.
func (s *MemoryStorage) Increment(key string, amount int, ttl time.Duration) (int64, error) {
	for {
		entryAny, ok := s.data.Load(key)
		if !ok {
			entry := &memoryEntry{
				value:     int64(amount),
				expiresAt: time.Now().Add(ttl),
			}
			//another go routing may created it
			actual, loaded := s.data.LoadOrStore(key, entry)
			if !loaded {
				return int64(amount), nil
			}
			entryAny = actual
		}

		entry, ok := entryAny.(*memoryEntry)
		if !ok {
			return 0, errors.New("invalid entry type")
		}
		//expiry
		if time.Now().After(entry.expiresAt) {
			newEntry := &memoryEntry{
				value:     int64(amount),
				expiresAt: time.Now().Add(ttl),
			}
			if s.data.CompareAndSwap(key, entry, newEntry) {
				return int64(amount), nil
			}
			continue // Retry
		}

		currentValue, ok := entry.value.(int64)
		if !ok {
			return 0, errors.New("value is not int64")
		}

		newValue := currentValue + int64(amount)
		newEntry := &memoryEntry{
			value:     newValue,
			expiresAt: entry.expiresAt,
		}

		if s.data.CompareAndSwap(key, entry, newEntry) {
			return newValue, nil
		}
		// else retry
	}
}
