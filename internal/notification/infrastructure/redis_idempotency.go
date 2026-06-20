package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type RedisIdempotencyStore struct {
	rdb *redis.Client
}

func NewRedisIdempotencyStore(rdb *redis.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{rdb: rdb}
}

func (s *RedisIdempotencyStore) CheckAndSet(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := s.rdb.SetNX(ctx, "notif:idem:"+key, "1", ttl).Result()
	if err != nil {
		return false, shared.ErrInternal.WithCause(err)
	}
	return ok, nil
}
