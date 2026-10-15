package internal

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

type RatelimiterManager struct {
	store Store

	totalRequests   int64
	allowedRequests int64
	deniedRequests  int64
}

func NewratelimiterManager() *RatelimiterManager {
	return NewRatelimiterManagerWithStore(NewMemoryStore())
}

func NewRatelimiterManagerWithStore(store Store) *RatelimiterManager {
	return &RatelimiterManager{
		store: store,
	}
}

func (rl *RatelimiterManager) Allow(
	identifier string,
	capacity, refillRate float64,
) (bool, float64) {
	atomic.AddInt64(&rl.totalRequests, 1)

	allowed, remaining := rl.store.Allow(identifier, capacity, refillRate)

	if allowed {
		atomic.AddInt64(&rl.allowedRequests, 1)
	} else {
		atomic.AddInt64(&rl.deniedRequests, 1)
	}

	return allowed, remaining
}

func (rl *RatelimiterManager) GetRemaining(identifier string) float64 {
	return rl.store.GetRemaining(identifier)
}

func (rl *RatelimiterManager) StartCleanup(
	cleanupInterval, maxIdle time.Duration,
) {
	ticker := time.NewTicker(cleanupInterval)

	go func() {
		log.Printf(
			"Cleanup started: interval=%v, maxIdle=%v",
			cleanupInterval,
			maxIdle,
		)
		for range ticker.C {
			rl.store.Cleanup(maxIdle)
		}
	}()
}

func (rl *RatelimiterManager) cleanup(maxIdle time.Duration) {
	rl.store.Cleanup(maxIdle)
}

func (rl *RatelimiterManager) GetStats() Stats {
	total := atomic.LoadInt64(&rl.totalRequests)
	denied := atomic.LoadInt64(&rl.deniedRequests)
	allowed := atomic.LoadInt64(&rl.allowedRequests)

	rate := 0.0
	if total > 0 {
		rate = float64(denied) / float64(total) * 100
	}

	perClient := rl.store.Snapshot()

	return Stats{
		Total:         total,
		Allowed:       allowed,
		Denied:        denied,
		RejectionRate: fmt.Sprintf("%.2f%%", rate),
		ActiveClients: len(perClient),
		PerClient:     perClient,
	}
}
