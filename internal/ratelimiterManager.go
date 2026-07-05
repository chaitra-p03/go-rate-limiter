package internal

import (
	"fmt"
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
	rl.mu.Unlock()

	bucket.mu.Lock()
	bucket.capacity = capacity
	bucket.refillRate = refillRate
	bucket.mu.Unlock()

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
	bucket, exists := rl.buckets[identifier]
	rl.mu.RUnlock()
	if !exists {
		return 0
	}
	
	return bucket.remaining()
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
	defer rl.mu.Unlock()
	for identifier, bucket := range rl.buckets {
		bucket.mu.Lock()
		idle := now.Sub(bucket.lastRefillTime) > maxIdle
		bucket.mu.Unlock()
		if idle {
			delete(rl.buckets, identifier)
		}
	}
}

func (rl *RatelimiterManager) GetStats() Stats {
    total := atomic.LoadInt64(&rl.totalRequests)
    denied := atomic.LoadInt64(&rl.deniedRequests)
    allowed := atomic.LoadInt64(&rl.allowedRequests)

    rate := 0.0
    if total > 0 {
        rate = float64(denied) / float64(total) * 100
    }

    rl.mu.RLock()
    clients := make(map[string]ClientStats, len(rl.buckets))
    for id, b := range rl.buckets {
	clients[id] = b.stats()
	}
    rl.mu.RUnlock()

    return Stats{
        Total: total,
        Allowed: allowed,
        Denied: denied,
        RejectionRate: fmt.Sprintf("%.2f%%", rate),
        ActiveClients: len(clients),
        PerClient: clients,
    }
}
