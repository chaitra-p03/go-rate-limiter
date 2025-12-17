package internal

import (
	"log"
	"sync"
	"time"
)

type RatelimiterManager struct {
	buckets map[string]*Bucket
	mu      sync.RWMutex
}

func NewratelimiterManager() *RatelimiterManager {
	return &RatelimiterManager{
		buckets: make(map[string]*Bucket),
	}
}

func (rl *RatelimiterManager) Allow(identifier string, capacity, refillRate float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, exists := rl.buckets[identifier]
	if !exists {
		bucket = &Bucket{
			tokens:         capacity,
			capacity:       capacity,
			refillRate:     refillRate,
			lastRefillTime: time.Now(),
		}
		rl.buckets[identifier] = bucket
	}
	return bucket.take(1)
}

func (rl *RatelimiterManager) GetRemaining(identifier string) float64 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket, exists := rl.buckets[identifier]
	if !exists {
		return 0
	}
	bucket.refill()
	return bucket.tokens
}

// cleanup goroutine
func (rl *RatelimiterManager) StartCleanup(cleanupInterval, maxIdle time.Duration) {
	ticker := time.NewTicker(cleanupInterval)

	go func() {
		log.Printf("Cleanup started: interval=%v, maxIdle=%v", cleanupInterval, maxIdle)
		for range ticker.C {
			rl.cleanup(maxIdle)
		}
	}()
}

// cleanup idle buckets
func (rl *RatelimiterManager) cleanup(maxIdle time.Duration) {
	now := time.Now()

	rl.mu.Lock()
	for identifier, bucker := range rl.buckets {

		if now.Sub(bucker.lastRefillTime) > maxIdle {
			delete(rl.buckets, identifier)
		}
	}
	rl.mu.Unlock()
}
