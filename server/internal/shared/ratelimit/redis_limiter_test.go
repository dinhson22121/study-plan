//go:build integration

package ratelimit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/pkg/redis"
)

func TestRedisLimiter_AllowsThenBlocksWithinWindow(t *testing.T) {
	url := os.Getenv("EDU_TEST_REDIS_URL")
	if url == "" {
		t.Skip("EDU_TEST_REDIS_URL not set")
	}
	rdb, err := redis.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })

	limiter := NewRedisLimiter(rdb, 3, time.Minute)
	key := "test:" + uuid.NewString()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		ok, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("allow %d: %v", i, err)
		}
		if !ok {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	ok, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("4th allow: %v", err)
	}
	if ok {
		t.Fatalf("4th request should be blocked")
	}
}

func TestRedisLimiter_WindowResets(t *testing.T) {
	url := os.Getenv("EDU_TEST_REDIS_URL")
	if url == "" {
		t.Skip("EDU_TEST_REDIS_URL not set")
	}
	rdb, err := redis.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })

	limiter := NewRedisLimiter(rdb, 1, time.Second)
	key := "test:" + uuid.NewString()
	ctx := context.Background()

	if ok, _ := limiter.Allow(ctx, key); !ok {
		t.Fatalf("first request should pass")
	}
	if ok, _ := limiter.Allow(ctx, key); ok {
		t.Fatalf("second request in window should block")
	}
	time.Sleep(1200 * time.Millisecond)
	if ok, _ := limiter.Allow(ctx, key); !ok {
		t.Fatalf("request after window reset should pass")
	}
}
