package internal

import (
	"sync"
	"testing"
	"time"
)

func TestNewBucketCreated(t *testing.T) {
	rl := NewratelimiterManager()
	allowed := rl.Allow("user1",10,1)
	if !allowed {
		t.Fatal("first request should always be allowed")
	}
}

func TestAllowedWithinLimit(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i<10; i++ {
		if !rl.Allow("user1",10,1) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestDeniedAfterLimitExceeded(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i<10; i++ {
		rl.Allow("user1",10,1)
	}
	if rl.Allow("user1",10,1) {
		t.Fatal("request should be denied after capacity exhausted")
	}
}

func TestTokenRefillAfterTime(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i<5; i++ {
		rl.Allow("user1",5,5) // refillRate=5
	}
	if rl.Allow("user1",5,5) {
		t.Fatal("should be denied immediately after drain")
	}
	time.Sleep(1 * time.Second) // 5×1 = 5 tokens back
	if !rl.Allow("user1",5,5) {
		t.Fatal("should be allowed after refill")
	}
}

func TestCleanupRemovesIdleBuckets(t *testing.T) {
	rl := NewratelimiterManager()
	rl.Allow("user1",10,1)

	time.Sleep(10*time.Millisecond)
	rl.cleanup(1*time.Millisecond) // anything idle >1ms gets deleted

	remaining := rl.GetRemaining("user1")
	if remaining != 0 {
		t.Fatalf("expected 0 after cleanup, got %v", remaining)
	}
}

func TestConcurrentRequests(t *testing.T) {
	rl := NewratelimiterManager()
	var wg sync.WaitGroup
	for i := 0; i<100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Allow("user1",100,10)
		}()
	}
	wg.Wait() // passes cleanly only if no race conditions
}

func TestIsolatedIdentifiers(t *testing.T) {
	rl := NewratelimiterManager()
	for i := 0; i<10; i++ {
		rl.Allow("user1",10,1) // drain user1
	}
	if !rl.Allow("user2",10,1) {
		t.Fatal("user2 should be unaffected by user1's bucket")
	}
}