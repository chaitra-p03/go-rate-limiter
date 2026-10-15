package internal

import (
	"sync"
	"time"
)

// Store describes the operations the rate limiter needs from storage.
type Store interface {
	Allow(identifier string, capacity, refillRate float64) (bool, float64)
	GetRemaining(identifier string) float64
	Cleanup(maxIdle time.Duration)
	Snapshot() map[string]ClientStats
}

// MemoryStore keeps each client's token bucket in this process.
type MemoryStore struct {
	buckets map[string]*Bucket
	mu      sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets: make(map[string]*Bucket),
	}
}

func (s *MemoryStore) Allow(
	identifier string,
	capacity, refillRate float64,
) (bool, float64) {
	s.mu.Lock()
	bucket, exists := s.buckets[identifier]

	if !exists {
		bucket = &Bucket{
			tokens:         capacity,
			capacity:       capacity,
			refillRate:     refillRate,
			lastRefillTime: time.Now(),
		}
		s.buckets[identifier] = bucket
	}
	s.mu.Unlock()

	bucket.mu.Lock()
	bucket.capacity = capacity
	bucket.refillRate = refillRate
	bucket.mu.Unlock()

	return bucket.take(1)
}

func (s *MemoryStore) GetRemaining(identifier string) float64 {
	s.mu.RLock()
	bucket, exists := s.buckets[identifier]
	s.mu.RUnlock()

	if !exists {
		return 0
	}

	return bucket.remaining()
}

func (s *MemoryStore) Cleanup(maxIdle time.Duration) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for identifier, bucket := range s.buckets {
		bucket.mu.Lock()
		idle := now.Sub(bucket.lastRefillTime) > maxIdle
		bucket.mu.Unlock()

		if idle {
			delete(s.buckets, identifier)
		}
	}
}

func (s *MemoryStore) Snapshot() map[string]ClientStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]ClientStats, len(s.buckets))
	for identifier, bucket := range s.buckets {
		result[identifier] = bucket.stats()
	}

	return result
}
