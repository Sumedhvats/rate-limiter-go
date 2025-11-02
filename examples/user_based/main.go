package main

// This example shows per-user rate limiting
// Each user gets their own rate limit bucket
//
// Authentication: X-User-ID header (simplified for demo)
// Real app: extract from JWT, session, etc.

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sumedhvats/rate-limiter-go/middleware"
	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func profileHandler(w http.ResponseWriter,r *http.Request){
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Missing X-User-ID header", http.StatusUnauthorized)
		return
	}
	response := map[string]string{
		"user":    userID,
		"message": "Profile data accessed successfully!",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
func main(){
	store:=storage.NewMemoryStorage()
	rateLimiter:=limiter.NewSlidingWindowLimiter(store,limiter.Config{
		Rate:10,
		Window:time.Minute,
	})
	userKeyFunc:= func(r *http.Request)string{
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
            return r.RemoteAddr
        }
        return "user:" + userID
	}
		onLimit := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Rate limit exceeded. Please try again later.",
		})
	}
	mux :=http.NewServeMux()
	mux.HandleFunc("/api/profile",profileHandler)
	handler := middleware.RateLimitMiddleware(middleware.Config{
		Limiter: rateLimiter,
		KeyFunc: userKeyFunc,
		OnLimit: onLimit,
	})(mux)
	log.Fatal(http.ListenAndServe(":8080", handler))

}