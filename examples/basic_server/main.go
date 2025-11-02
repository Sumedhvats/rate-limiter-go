package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sumedhvats/rate-limiter-go/middleware"
	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
	"github.com/sumedhvats/rate-limiter-go/pkg/storage"
)

func main() {
	config := limiter.Config{
		Rate:   10,
		Window: 1 * time.Minute,
	}
	store := storage.NewMemoryStorage()
	limiter := limiter.NewSlidingWindowLimiter(store, config)
	mux := http.NewServeMux()

	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello from rate-limited API!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	handler := middleware.RateLimitMiddleware(middleware.Config{
		Limiter: limiter,
	})(mux)

	log.Println("Server starting on :8080")
	log.Println("Try: curl <http://localhost:8080/api/data>")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
