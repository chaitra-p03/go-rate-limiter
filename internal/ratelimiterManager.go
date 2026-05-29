package internal

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type RatelimiterManager struct {
	buckets map[string]*Bucket
	mu sync.RWMutex

	totalRequests int64
	allowedRequests int64
	deniedRequests int64
}

func NewratelimiterManager() *RatelimiterManager {
	return &RatelimiterManager{
		buckets: make(map[string]*Bucket),
	}
}

func (rl *RatelimiterManager) Allow(identifier string, capacity, refillRate float64) bool {
	atomic.AddInt64(&rl.totalRequests, 1)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[identifier]
	if !exists {
		bucket = &Bucket{
			tokens: capacity,
			capacity: capacity,
			refillRate: refillRate,
			lastRefillTime: time.Now(),
		}
		rl.buckets[identifier] = bucket
	}
	allowed := bucket.take(1)
	if allowed {
		atomic.AddInt64(&rl.allowedRequests, 1)
	} else {
		atomic.AddInt64(&rl.deniedRequests, 1)
	}
	return allowed
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
	for identifier, bucket := range rl.buckets {
		if now.Sub(bucket.lastRefillTime) > maxIdle {
			delete(rl.buckets, identifier)
		}
	}
	rl.mu.Unlock()
}

func (r1 *RatelimiterManager) GetStats() Stats {
	return Stats{
		Total:   atomic.LoadInt64(&r1.totalRequests),
		Allowed: atomic.LoadInt64(&r1.allowedRequests),
		Denied:  atomic.LoadInt64(&r1.deniedRequests),
	}
}
