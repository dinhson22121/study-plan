package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

type RedisRefreshStore struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewRedisRefreshStore(rdb *redis.Client, ttl time.Duration) *RedisRefreshStore {
	return &RedisRefreshStore{rdb: rdb, ttl: ttl}
}

func (s *RedisRefreshStore) key(userID, jti string) string {
	return fmt.Sprintf("auth:refresh:%s:%s", userID, jti)
}

func (s *RedisRefreshStore) Save(ctx context.Context, userID, jti string) error {
	if err := s.rdb.Set(ctx, s.key(userID, jti), "1", s.ttl).Err(); err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *RedisRefreshStore) Exists(ctx context.Context, userID, jti string) (bool, error) {
	n, err := s.rdb.Exists(ctx, s.key(userID, jti)).Result()
	if err != nil {
		return false, domain.ErrInternal.WithCause(err)
	}
	return n > 0, nil
}

func (s *RedisRefreshStore) Delete(ctx context.Context, userID, jti string) error {
	if err := s.rdb.Del(ctx, s.key(userID, jti)).Err(); err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}
