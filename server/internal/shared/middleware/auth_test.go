package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

func init() { gin.SetMode(gin.TestMode) }

func newRouter(validate TokenValidator) *gin.Engine {
	r := gin.New()
	r.GET("/me", Auth(validate), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": UserIDFrom(c), "role": RoleFrom(c)})
	})
	r.GET("/admin", Auth(validate), RequireRole(RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func validatorReturning(userID, role string, err error) TokenValidator {
	return func(context.Context, string) (*Claims, error) {
		if err != nil {
			return nil, err
		}
		return &Claims{UserID: userID, Role: role}, nil
	}
}

func TestAuth_RejectsMissingHeader(t *testing.T) {
	r := newRouter(validatorReturning("u1", RoleStudent, nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuth_RejectsMalformedHeader(t *testing.T) {
	r := newRouter(validatorReturning("u1", RoleStudent, nil))
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Token abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuth_RejectsInvalidToken(t *testing.T) {
	r := newRouter(validatorReturning("", "", domain.ErrUnauthorized))
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer bad")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuth_AllowsValidTokenAndSetsContext(t *testing.T) {
	r := newRouter(validatorReturning("u1", RoleStudent, nil))
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer good")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestRequireRole_ForbidsWrongRole(t *testing.T) {
	r := newRouter(validatorReturning("u1", RoleStudent, nil))
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer good")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestRequireRole_AllowsCorrectRole(t *testing.T) {
	r := newRouter(validatorReturning("admin1", RoleAdmin, nil))
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer good")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}
