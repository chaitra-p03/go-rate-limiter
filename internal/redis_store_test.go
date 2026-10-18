package internal

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisStoresShareBucket(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	store1 := &RedisStore{
		client:  client,
		prefix:  "rate-limit:",
		idleTTL: time.Minute,
	}
	store2 := &RedisStore{
		client:  client,
		prefix:  "rate-limit:",
		idleTTL: time.Minute,
	}

	ctx := context.Background()

	for range 2 {
		allowed, _, err := store1.Allow(ctx, "client-1", 2, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatal("request should be allowed")
		}
	}

	allowed, _, err := store2.Allow(ctx, "client-1", 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("third request should be denied across the two stores")
	}
}