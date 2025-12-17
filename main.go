package main

import (
	"encoding/json"
	"log"
	"net/http"
	"rateLimiter/internal"
	"time"
)

var limiter *internal.RatelimiterManager

func main() {
	limiter = internal.NewratelimiterManager()
	// cleanup - runs every 1 min, removes buckets idle for >10mins
	limiter.StartCleanup(1*time.Minute, 10*time.Minute)

	http.HandleFunc("/check", handleCheck)
	http.HandleFunc("/stats", statsHandler)

	log.Println("Rate limiter starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req internal.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Identifier == "" {
		http.Error(w, "identifier required", http.StatusBadRequest)
		return
	}
	if req.Capacity <= 0 || req.RefillRate <= 0 {
		http.Error(w, "capacity and refill rate must be positive", http.StatusBadRequest)
		return
	}

	allowed := limiter.Allow(req.Identifier, req.Capacity, req.RefillRate)
	remaining := limiter.GetRemaining(req.Identifier)

	response := internal.CheckResponse{
		Allowed:   allowed,
		Remaining: remaining,
		Limit:     req.Capacity,
	}

	if !allowed {
		retryAfter := 1.0 / req.RefillRate
		response.RetryAfter = retryAfter
	}
	w.Header().Set("Content-Type", "application/json")
	if !allowed {
		w.WriteHeader(http.StatusTooManyRequests)
	}
	_ = json.NewEncoder(w).Encode(response)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := limiter.GetStats()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
