// internal/middleware.go
package internal

import (
	"net"
	"net/http")

func (rl *RatelimiterManager) Middleware(capacity, refillRate float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err!=nil {
				ip=r.RemoteAddr
			}

			if !rl.Allow(ip, capacity, refillRate) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}