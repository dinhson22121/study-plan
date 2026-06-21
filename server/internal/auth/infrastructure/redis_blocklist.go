package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

type RedisBlocklist struct {
	rdb *redis.Client
}

func NewRedisBlocklist(rdb *redis.Client) *RedisBlocklist {
	return &RedisBlocklist{rdb: rdb}
}

func (b *RedisBlocklist) key(jti string) string {
	return "auth:revoked:" + jti
}

func (b *RedisBlocklist) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if err := b.rdb.Set(ctx, b.key(jti), "1", ttl).Err(); err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}

func (b *RedisBlocklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := b.rdb.Exists(ctx, b.key(jti)).Result()
	if err != nil {
		return false, domain.ErrInternal.WithCause(err)
	}
	return n > 0, nil
}
