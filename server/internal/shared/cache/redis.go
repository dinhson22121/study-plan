package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{rdb: rdb, prefix: "cache:"}
}

func (c *RedisCache) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	b, err := c.rdb.Get(ctx, c.prefix+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *RedisCache) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, c.prefix+key, b, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	full := make([]string, len(keys))
	for i, k := range keys {
		full[i] = c.prefix + k
	}
	return c.rdb.Del(ctx, full...).Err()
}
