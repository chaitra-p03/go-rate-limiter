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
	// cleanup - runs every 1 min and cleans those idle for >10 mins
	limiter.StartCleanup(1*time.Minute, 10*time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("/stats", statsHandler)
	limitedDataHandler := limiter.Middleware(10, 1)(http.HandlerFunc(dataHandler))
	mux.Handle("/api/data", limitedDataHandler)

	log.Println("Rate limiter starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := limiter.GetStats()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "ok"}`))
}