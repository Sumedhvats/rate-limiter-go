package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

type UserTier string

const (
	TierFree    UserTier = "free"
	TierPremium UserTier = "premium"
	TierAdmin   UserTier = "admin"
)

func main() {
	store := storage.NewRedisStorage("localhost:6379")
	limiters := map[UserTier]limiter.Limiter{
        TierFree: limiter.NewSlidingWindowLimiter(store, limiter.Config{
            Rate:   100,
            Window: 1 * time.Hour,
        }),
        TierPremium: limiter.NewSlidingWindowLimiter(store, limiter.Config{
            Rate:   1000,
            Window: 1 * time.Hour,
        }),
        TierAdmin: nil,
    }
	tierMiddleware:= func(next http.Handler) http.Handler{
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tier:=getUserTier(r)
			 if tier == TierAdmin {
                next.ServeHTTP(w, r)
                return
            }
			lim := limiters[tier]
			userID := r.Header.Get("X-User-ID")

            allowed, _ := lim.Allow(userID)
            if !allowed {
                http.Error(w, "Rate limit exceeded", 429)
                return
            }

            next.ServeHTTP(w, r)
        })
	}
    mux:=http.NewServeMux()
    mux.HandleFunc("/api/data",func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{
			 "message": "Hello from rate-limited API!",
            "time":    time.Now().Format(time.RFC3339),
		})
    })

    handler:=tierMiddleware(mux)
    log.Fatal(http.ListenAndServe(":8080", handler))


}


// Replace this with JWT or session verification.
func getUserTier(r *http.Request) UserTier {
    return UserTier(r.Header.Get("X-User-Tier"))
}
