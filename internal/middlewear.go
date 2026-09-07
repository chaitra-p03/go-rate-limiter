package internal

import (
	"net/http"
)

func (rl *RatelimiterManager) Middleware(capacity, refillRate float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := ResolveClient(r)
			allowed, _ := rl.Allow(clientKey, capacity, refillRate)
			if !allowed {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}