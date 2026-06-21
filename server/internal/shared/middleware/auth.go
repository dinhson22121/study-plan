package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
)

const (
	ctxUserID = "auth.user_id"
	ctxRole   = "auth.role"
)

const (
	RoleStudent = "STUDENT"
	RoleAdmin   = "ADMIN"
)

type Claims struct {
	UserID string
	Role   string
}

type TokenValidator func(ctx context.Context, token string) (*Claims, error)

func Auth(validate TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := bearerToken(c)
		if err != nil {
			httpx.Fail(c, err)
			c.Abort()
			return
		}
		claims, err := validate(c.Request.Context(), token)
		if err != nil {
			httpx.Fail(c, domain.ErrUnauthorized.WithCause(err))
			c.Abort()
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if RoleFrom(c) != role {
			httpx.Fail(c, domain.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func UserIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxUserID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func RoleFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxRole); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func bearerToken(c *gin.Context) (string, error) {
	h := c.GetHeader("Authorization")
	if h == "" {
		return "", domain.ErrUnauthorized.WithMessage("missing Authorization header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", domain.ErrUnauthorized.WithMessage("malformed Authorization header")
	}
	return parts[1], nil
}
