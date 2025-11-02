package middleware

import (
	"encoding/json"
	"strings"

	"fmt"
	"net"
	"net/http"

	"github.com/sumedhvats/rate-limiter-go/pkg/limiter"
)

type Config struct {
	Limiter limiter.Limiter
	KeyFunc func(*http.Request) string
	OnLimit func(http.ResponseWriter, *http.Request)
}

func RateLimitMiddleware(cfg Config) func(http.Handler) http.Handler {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = DefaultKeyFunc
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.KeyFunc(r)
			if key == "" {
				http.Error(w, "Unable to determine client IP", http.StatusBadRequest)
				return
			}
			allowed, err := cfg.Limiter.Allow(key)
			if err != nil {
				http.Error(w, "Internal Server Error", 500)
				return
			}

			if !allowed {
				if cfg.OnLimit != nil {
					cfg.OnLimit(w, r)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded. Try again later.",
				})
				return
			}

			stats, _ := cfg.Limiter.GetStats(key)
			if stats != nil {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprint(stats.Limit))
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprint(stats.Remaining))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprint(stats.ResetAt.Unix()))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func DefaultKeyFunc(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, part := range parts {
			ip := strings.TrimSpace(part)
			if ip != "" {
				if norm := normalizeIP(ip); norm != "" {
					return norm
				}
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if norm := normalizeIP(host); norm != "" {
		return norm
	}
	return ""
}
func normalizeIP(ipStr string) string {
	if i := strings.IndexByte(ipStr, '%'); i >= 0 {
		ipStr = ipStr[:i]
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4.String()
	}
	return ip.String()
}
