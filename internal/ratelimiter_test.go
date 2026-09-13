package internal

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewBucketCreated(t *testing.T) {
	rl := NewratelimiterManager()
	allowed, _ := rl.Allow("user1", 10, 1)
	if !allowed {
		t.Fatal("first request should always be allowed")
	}
}

func TestAllowedWithinLimit(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i < 10; i++ {
		allowed, _ := rl.Allow("user1", 10, 1)
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestDeniedAfterLimitExceeded(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i < 10; i++ {
		rl.Allow("user1", 10, 1)
	}
	allowed, _ := rl.Allow("user1", 10, 1)
	if allowed {
		t.Fatal("request should be denied after capacity exhausted")
	}
}

func TestTokenRefillAfterTime(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i < 5; i++ {
		rl.Allow("user1", 5, 5) // refillRate=5
	}
	allowed, _ := rl.Allow("user1", 5, 5)
	if allowed {
		t.Fatal("should be denied immediately after drain")
	}
	time.Sleep(1 * time.Second) // 5×1 = 5 tokens back
	allowed, _ = rl.Allow("user1", 5, 5)
	if !allowed {
		t.Fatal("should be allowed after refill")
	}
}

func TestCleanupRemovesIdleBuckets(t *testing.T) {
	rl := NewratelimiterManager()
	rl.Allow("user1", 10, 1)

	time.Sleep(10 * time.Millisecond)
	rl.cleanup(1 * time.Millisecond) // anything idle >1ms gets deleted

	remaining := rl.GetRemaining("user1")
	if remaining != 0 {
		t.Fatalf("expected 0 after cleanup, got %v", remaining)
	}
}

func TestConcurrentRequestsRespectCapacity(t *testing.T) {
	rl := NewratelimiterManager()
	var wg sync.WaitGroup
	var allowedCount int64
	for i := 0; i < 150; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _ := rl.Allow("user1", 100, 0)
			if allowed {
				atomic.AddInt64(&allowedCount, 1)
			}
		}()
	}

	wg.Wait()
	if allowedCount != 100 {
		t.Fatalf("expected 100 allowed requests, got %d", allowedCount)
	}
	deniedCount := int64(150) - allowedCount
	if deniedCount != 50 {
		t.Fatalf("expected 50 denied requests, got %d", deniedCount)
	}
}

func TestIsolatedIdentifiers(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i < 10; i++ {
		rl.Allow("user1", 10, 1) // drain user1
	}
	allowed, _ := rl.Allow("user2", 10, 1)
	if !allowed {
		t.Fatal("user2 should be unaffected by user1's bucket")
	}
}
