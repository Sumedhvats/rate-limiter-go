package storage

import (
	"errors"

	"sync"
	"time"
)
type MemoryStorage struct{
	data sync.Map
	cleanup *time.Ticker
}
type memoryEntry struct{
	value interface{}
	expiresAt time.Time
}
func NewMemoryStorage() *MemoryStorage{
	s:=&MemoryStorage{
		cleanup: time.NewTicker(1*time.Minute),
	}
	go s.cleanupExpired()
	return s
}
func (s *MemoryStorage) cleanupExpired(){
	for range s.cleanup.C{
		now :=time.Now()
		s.data.Range(func(key, value interface{}) bool {
			entry:=value.(memoryEntry)
			if now.After(entry.expiresAt){
				s.data.Delete(key)
			}
			return true
		})
		
	}
}
func (s *MemoryStorage) Get(key string)(interface{},error){
	val,ok := s.data.Load(key)
	if !ok{
		return nil,nil
	}
	entry:=val.(*memoryEntry)
   if time.Now().After(entry.expiresAt) {
        s.data.Delete(key)
        return nil, nil
    }
	return entry.value, nil
}
func (s *MemoryStorage) Set(key string,value interface{}, ttl time.Duration)error{
	now:= time.Now()
	entry:=&memoryEntry{
		value: value,
		expiresAt: now.Add(ttl),
	}
   s.data.Store(key,entry)
   return nil
}

func (s *MemoryStorage) Delete(key string){
s.data.Delete(key)
}

func (s *MemoryStorage) Increment(key string, amount int) (int64, error) {
	for {
		entryAny, ok := s.data.Load(key)
		if !ok {
			entry := &memoryEntry{
				value:     int64(amount),
				expiresAt: time.Now().Add(10 * time.Minute),
			}
			s.data.Store(key, entry)
			return int64(amount), nil
		}

		entry, ok := entryAny.(*memoryEntry)
		if !ok {
			return 0, errors.New("invalid entry type")
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

		swapped := s.data.CompareAndSwap(key, entry, newEntry)
		if swapped {
			return newValue, nil
		}
		// else retry
	}
}