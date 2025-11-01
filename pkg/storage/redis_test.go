package storage

import (
	"context"

	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisStorage_GetSet(t *testing.T) {

	store := NewRedisMemory("127.0.0.1:6379")

	err := store.Set("123.234.15.1", "value1", 1*time.Minute)
	assert.NoError(t, err)

	val, err := store.Get("123.234.15.1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}
func TestExpirationRedis(t *testing.T) {

	store := NewRedisMemory("127.0.0.1:6379")

	ttl := 100 * time.Millisecond
	err := store.Set("tempKey", "tempValue", ttl)
	assert.NoError(t, err)

	val, err := store.Get("tempKey")
	assert.NoError(t, err)
	assert.Equal(t, "tempValue", val)

	time.Sleep(ttl + 100*time.Millisecond)
	val, err = store.Get("tempKey")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestMemoryStorate_IncrementRedis(t *testing.T) {

	store := NewRedisMemory("127.0.0.1:6379")
	store.client.FlushAll(context.Background())
	val, err := store.Increment("counter", 5, time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), val)

	val, err = store.Increment("counter", 3, time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), val)

	got, err := store.Get("counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(8), got.(int64))

}

func TestConcurrentAccessRedis(t *testing.T) {
	store := NewRedisMemory("127.0.0.1:6379")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Increment("temp", 100, time.Minute)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	currVal, err := store.Get("temp")
	assert.NoError(t, err)
	assert.Equal(t, currVal, int64(10000))
}
