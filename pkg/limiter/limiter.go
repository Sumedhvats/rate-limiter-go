package limiter

import "time"
type stats struct{
	Limit int
	Remaining int
	ResetAt time.Time
}
type Limiter interface{
	Allow(key string)(bool,error)
	AllowN(key string ,n int)(bool,error)
	Reset(key string) error
	GetStats(key string)(*stats,error)
}
type config struct{
	Rate int
	Window time.Duration
	Burst int
}

type Storage interface{
	Get(key string)(interface{},error)
	Set(key string , value interface{}, ttl time.Duration)error
	Increment(key string, value int)(int64,error)
	Delete(key string)error
}