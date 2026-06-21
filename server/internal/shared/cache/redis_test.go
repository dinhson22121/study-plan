//go:build integration

package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/pkg/redis"
)

func testCache(t *testing.T) *RedisCache {
	t.Helper()
	url := os.Getenv("EDU_TEST_REDIS_URL")
	if url == "" {
		t.Skip("EDU_TEST_REDIS_URL not set")
	}
	rdb, err := redis.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return NewRedisCache(rdb)
}

type payload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestRedisCache_SetGetDelete(t *testing.T) {
	c := testCache(t)
	ctx := context.Background()
	key := "test:" + uuid.NewString()

	var miss payload
	if found, err := c.GetJSON(ctx, key, &miss); err != nil || found {
		t.Fatalf("expected miss (found=%v err=%v)", found, err)
	}

	if err := c.SetJSON(ctx, key, payload{Name: "x", Count: 3}, time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}
	var got payload
	if found, err := c.GetJSON(ctx, key, &got); err != nil || !found {
		t.Fatalf("expected hit (found=%v err=%v)", found, err)
	}
	if got.Name != "x" || got.Count != 3 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}

	if err := c.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if found, _ := c.GetJSON(ctx, key, &got); found {
		t.Fatalf("expected miss after delete")
	}
}

func TestRedisCache_TTLExpires(t *testing.T) {
	c := testCache(t)
	ctx := context.Background()
	key := "test:" + uuid.NewString()

	if err := c.SetJSON(ctx, key, payload{Name: "y"}, time.Second); err != nil {
		t.Fatalf("set: %v", err)
	}
	time.Sleep(1200 * time.Millisecond)
	var got payload
	if found, _ := c.GetJSON(ctx, key, &got); found {
		t.Fatalf("expected expiry after ttl")
	}
}
