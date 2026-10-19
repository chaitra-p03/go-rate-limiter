package internal

import (
	"math"
	"net/http"
	"strconv"
)

func (rl *RatelimiterManager) Middleware(capacity, refillRate float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := ResolveClient(r)
			allowed, remaining, err := rl.AllowContext(r.Context(),clientKey,capacity,refillRate)
			if err != nil {
				http.Error(w, "rate limiter unavailable", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set(
				"RateLimit-Limit",
				strconv.FormatFloat(capacity, 'f', -1, 64),
			)

			w.Header().Set(
				"RateLimit-Remaining",
				strconv.FormatFloat(remaining, 'f', -1, 64),
			)
			if !allowed {
				if refillRate > 0 {
					retryAfter := math.Ceil((1 - remaining) / refillRate)
					w.Header().Set(
						"Retry-After",
						strconv.FormatFloat(retryAfter, 'f', 0, 64),
					)
				}
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
			next.ServeHTTP(w, r)
		})
	}
}