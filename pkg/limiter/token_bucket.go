package limiter

import (
	"fmt"
	"time"
)

type tokenBucket struct{
	tokens float64
	lastRefilTime time.Time
	capacity int
	refillRate float64
}
type TokenBucketLimiter struct{
	storage Storage
	config config
}
func NewTokenBucketLimiter(store Storage, cfg config)(*TokenBucketLimiter){
	return &TokenBucketLimiter{
		storage: store,
		config: cfg,
	}
}

func (t *TokenBucketLimiter)AllowN(key string,n int)(bool,error){
	now:=time.Now()
	data,err:= t.storage.Get(key)
	if err!=nil{
		return false,err
	}
	var bucket *tokenBucket
	if data ==nil{
		bucket=&tokenBucket{
			tokens: float64(t.config.Burst),
			lastRefilTime:now ,
			capacity: t.config.Burst,
			refillRate: float64(t.config.Rate)/float64(t.config.Window.Seconds()),
		}
	}else{
		bucket=data.(*tokenBucket)
		elapsed:= now.Sub(bucket.lastRefilTime).Seconds()
		tokenAdded:=elapsed*bucket.refillRate
		bucket.tokens = min(bucket.tokens+tokenAdded, float64(bucket.capacity))
		bucket.lastRefilTime=now
	}
	if(bucket.tokens>float64(n)){
		bucket.tokens-=float64(n)
		t.storage.Set(key,bucket,t.config.Window)
		return true,nil
	}
	t.storage.Set(key,bucket,t.config.Window)
	return false,nil
}
func (t *TokenBucketLimiter) Allow(key string)(bool,error){
	return t.AllowN(key,1)
}
func (t *TokenBucketLimiter)Reset(key string) error{
	return t.storage.Delete(key)
}
func (t *TokenBucketLimiter) GetStats(key string) (*stats, error) {

	data,_:=t.storage.Get(key)
	var bucket *tokenBucket
	if(data==nil){
		return &stats{},fmt.Errorf("not found key")
	}else{
		bucket=data.(*tokenBucket)
	bucketStats:=&stats{
		Limit: t.config.Burst,
		Remaining: int(bucket.tokens),
		ResetAt: bucket.lastRefilTime.Add(t.config.Window),
	}
	return bucketStats,nil
}
}