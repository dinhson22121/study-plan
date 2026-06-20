// Package middleware holds cross-cutting Gin middleware: authentication,
// authorization, request logging, and panic recovery.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
)

// contextKey constants for values stashed on the Gin context by Auth.
const (
	ctxUserID = "auth.user_id"
	ctxRole   = "auth.role"
)

// Roles recognized by RequireRole.
const (
	RoleStudent = "STUDENT"
	RoleAdmin   = "ADMIN"
)

// Claims is the minimal identity extracted from a validated access token.
type Claims struct {
	UserID string
	Role   string
}

// TokenValidator validates a raw access token and returns its claims. The auth
// module supplies the concrete implementation, keeping this package decoupled
// from JWT internals.
type TokenValidator func(token string) (*Claims, error)

// Auth returns middleware that requires a valid Bearer access token. On success
// it stashes the user id and role on the context; on failure it aborts with 401.
func Auth(validate TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := bearerToken(c)
		if err != nil {
			httpx.Fail(c, err)
			c.Abort()
			return
		}
		claims, err := validate(token)
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

// RequireRole returns middleware that enforces a specific role. It must run
// after Auth, which populates the role on the context.
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

// UserIDFrom returns the authenticated user id, or "" if unauthenticated.
func UserIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxUserID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RoleFrom returns the authenticated user's role, or "" if unauthenticated.
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
