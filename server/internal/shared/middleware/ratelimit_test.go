package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeLimiter struct {
	allowed  bool
	err      error
	gotKey   string
	gotCalls int
}

func (f *fakeLimiter) Allow(_ context.Context, key string) (bool, error) {
	f.gotCalls++
	f.gotKey = key
	return f.allowed, f.err
}

func runWith(mw gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/t", mw, func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.RemoteAddr = "203.0.113.5:1234"
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	lim := &fakeLimiter{allowed: true}
	w := runWith(RateLimit(lim, "auth"))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if lim.gotKey != "auth:203.0.113.5" {
		t.Fatalf("unexpected key %q", lim.gotKey)
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	w := runWith(RateLimit(&fakeLimiter{allowed: false}, "auth"))
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", w.Code)
	}
}

func TestRateLimit_FailsOpenOnLimiterError(t *testing.T) {
	w := runWith(RateLimit(&fakeLimiter{err: errors.New("redis down")}, "auth"))
	if w.Code != http.StatusOK {
		t.Fatalf("want fail-open 200, got %d", w.Code)
	}
}

func TestRateLimit_NilLimiterIsNoop(t *testing.T) {
	w := runWith(RateLimit(nil, "auth"))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 when limiter nil, got %d", w.Code)
	}
}
