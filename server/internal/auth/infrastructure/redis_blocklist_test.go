//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/pkg/redis"
)

func testBlocklist(t *testing.T) *RedisBlocklist {
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
	return NewRedisBlocklist(rdb)
}

func TestRedisBlocklist_RevokeAndCheck(t *testing.T) {
	bl := testBlocklist(t)
	ctx := context.Background()
	jti := uuid.NewString()

	if revoked, err := bl.IsRevoked(ctx, jti); err != nil || revoked {
		t.Fatalf("fresh jti should not be revoked (revoked=%v err=%v)", revoked, err)
	}
	if err := bl.Revoke(ctx, jti, time.Minute); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if revoked, err := bl.IsRevoked(ctx, jti); err != nil || !revoked {
		t.Fatalf("jti should be revoked (revoked=%v err=%v)", revoked, err)
	}
}

func TestRedisBlocklist_TTLExpires(t *testing.T) {
	bl := testBlocklist(t)
	ctx := context.Background()
	jti := uuid.NewString()

	if err := bl.Revoke(ctx, jti, time.Second); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	time.Sleep(1200 * time.Millisecond)
	if revoked, err := bl.IsRevoked(ctx, jti); err != nil || revoked {
		t.Fatalf("revocation should expire after ttl (revoked=%v err=%v)", revoked, err)
	}
}
