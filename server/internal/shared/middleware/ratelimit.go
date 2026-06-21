package middleware

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

func RateLimit(limiter RateLimiter, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}
		key := scope + ":" + c.ClientIP()
		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			httpx.Fail(c, domain.ErrTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}
