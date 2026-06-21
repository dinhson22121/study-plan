package cache

import (
	"context"
	"time"
)

// Cache is a small JSON key/value cache (Redis-backed in production). Reads are
// best-effort: callers fall through to the source of truth on a miss or error.
type Cache interface {
	GetJSON(ctx context.Context, key string, dest any) (found bool, err error)
	SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

// Aside implements the cache-aside pattern: return the cached value on a hit,
// otherwise load from source, populate the cache (best-effort) and return it.
// A nil cache (or any cache error) transparently falls back to load().
func Aside[T any](
	ctx context.Context,
	c Cache,
	key string,
	ttl time.Duration,
	load func() (T, error),
) (T, error) {
	var zero T
	if c != nil {
		var cached T
		if found, err := c.GetJSON(ctx, key, &cached); err == nil && found {
			return cached, nil
		}
	}
	val, err := load()
	if err != nil {
		return zero, err
	}
	if c != nil {
		_ = c.SetJSON(ctx, key, val, ttl)
	}
	return val, nil
}

// Invalidate deletes keys best-effort; a nil cache is a no-op.
func Invalidate(ctx context.Context, c Cache, keys ...string) {
	if c == nil || len(keys) == 0 {
		return
	}
	_ = c.Delete(ctx, keys...)
}
