package main

import (
	"encoding/json"
	"log"
	"net/http"
	"rateLimiter/internal"
	"time"
)

func main() {
	limiter := internal.NewratelimiterManager()
	limiter.StartCleanup(time.Minute, 10*time.Minute)

	log.Println("Rate limiter starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", newRouter(limiter, 10, 1)))
}

// Passing in the limiter and policy lets tests build a router
// without starting a real server.
func newRouter(
	limiter *internal.RatelimiterManager,
	capacity float64,
	refillRate float64,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(limiter.GetStats())
	})

	limitedDataHandler := limiter.Middleware(
		capacity,
		refillRate,
	)(http.HandlerFunc(dataHandler))

	mux.Handle("/api/data", limitedDataHandler)

	return mux
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"message": "ok"}`))
}
