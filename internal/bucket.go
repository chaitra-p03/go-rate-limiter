package internal

import "time"

type Bucket struct {
	tokens         float64
	capacity       float64
	refillRate     float64
	lastRefillTime time.Time
}

func (b *Bucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefillTime).Seconds()
	b.tokens += b.refillRate * elapsed
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefillTime = now
}

func (b *Bucket) take(n float64) bool {
	b.refill()
	if b.tokens >= n {
		b.tokens = b.tokens - n
		return true
	}
	return false
}
