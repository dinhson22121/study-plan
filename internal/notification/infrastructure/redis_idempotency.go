package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// RedisIdempotencyStore implements domain.IdempotencyStore using Redis SETNX,
// so the first caller for a key wins and later duplicates are detected for the
// TTL window (PRD: 24h).
type RedisIdempotencyStore struct {
	rdb *redis.Client
}

// NewRedisIdempotencyStore builds the store.
func NewRedisIdempotencyStore(rdb *redis.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{rdb: rdb}
}

// CheckAndSet atomically records the key with a TTL and reports whether it was
// newly set (true = first time seen, false = duplicate).
func (s *RedisIdempotencyStore) CheckAndSet(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := s.rdb.SetNX(ctx, "notif:idem:"+key, "1", ttl).Result()
	if err != nil {
		return false, shared.ErrInternal.WithCause(err)
	}
	return ok, nil
}
