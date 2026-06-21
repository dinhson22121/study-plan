package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// allowScript atomically increments the window counter and, on the first hit,
// sets its TTL. Doing both in one Lua call avoids the INCR/EXPIRE race where a
// crash or dropped connection between the two commands would leave the counter
// key with no expiry — permanently banning the client until Redis restarts.
var allowScript = redis.NewScript(`
local n = redis.call("INCR", KEYS[1])
if n == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return n
`)

type RedisLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func NewRedisLimiter(rdb *redis.Client, limit int, window time.Duration) *RedisLimiter {
	return &RedisLimiter{rdb: rdb, limit: limit, window: window}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rkey := "ratelimit:" + key
	count, err := allowScript.Run(ctx, l.rdb, []string{rkey}, int64(l.window.Seconds())).Int64()
	if err != nil {
		return false, err
	}
	return count <= int64(l.limit), nil
}
