package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sumedhvats/rate-limiter-go/middleware"
	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)
func TestRateLimitMiddleware(t *testing.T) {
	store:= storage.NewMemoryStorage()
	limiter := limiter.NewSlidingWindowLimiter(store, limiter.Config{
        Rate:   5,
        Window: 1 * time.Minute,
    })
	handler:=http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	wrapped:=middleware.RateLimitMiddleware(middleware.Config{
		Limiter: limiter,
	})(handler)
	for i := 0; i < 5; i++ {
        req := httptest.NewRequest("GET", "/api/test", nil)
        req.RemoteAddr = "192.168.1.1:12345"
        rr := httptest.NewRecorder()

        wrapped.ServeHTTP(rr, req)

        assert.Equal(t, 200, rr.Code)
        assert.Equal(t, "5", rr.Header().Get("X-RateLimit-Limit"))
    }
	req:=httptest.NewRequest("GET","/api/test",nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr:=httptest.NewRecorder()

 wrapped.ServeHTTP(rr, req)
 assert.Equal(t, 429, rr.Code)
    assert.Contains(t, rr.Body.String(), "Rate limit exceeded")
}


func TestCustomKeyFunc(t *testing.T) {
	userKeyFunc:= func(r *http.Request) string{
		return r.Header.Get("X-USER-ID")
	}
	store:= storage.NewMemoryStorage()
	limiter := limiter.NewSlidingWindowLimiter(store, limiter.Config{
        Rate:   5,
        Window: 1 * time.Minute,
    })
	cfg := middleware.Config{
        Limiter: limiter,
        KeyFunc: userKeyFunc,
    }
	//user-1
	for i :=0 ;i<5 ; i++ {
	req:=httptest.NewRequest("GET","/api/test",nil)
	req.Header.Add("X-USER-ID","user-1")
	rr:=httptest.NewRecorder()
	wraped:=middleware.RateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	wraped.ServeHTTP(rr,req)
	}



	//user-2
	for i :=0 ;i<5 ; i++ {
	req:=httptest.NewRequest("GET","/api/test",nil)
	req.Header.Add("X-USER-ID","user-2")
	rr:=httptest.NewRecorder()
	wraped:=middleware.RateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	wraped.ServeHTTP(rr,req)
	assert.Equal(t,200,rr.Code)
	}

}
