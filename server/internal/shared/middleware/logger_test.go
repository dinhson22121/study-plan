package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestLogger_AssignsAndPropagatesCorrelationID(t *testing.T) {
	r := gin.New()
	r.Use(Logger(zap.NewNop()))
	var seen string
	r.GET("/", func(c *gin.Context) {
		seen = CorrelationIDFrom(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if seen == "" {
		t.Fatalf("expected a correlation id to be assigned")
	}
	if w.Header().Get(correlationHeader) != seen {
		t.Fatalf("correlation id not echoed in response header")
	}
}

func TestLogger_HonorsInboundCorrelationID(t *testing.T) {
	r := gin.New()
	r.Use(Logger(zap.NewNop()))
	var seen string
	r.GET("/", func(c *gin.Context) { seen = CorrelationIDFrom(c) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(correlationHeader, "abc-123")
	r.ServeHTTP(httptest.NewRecorder(), req)

	if seen != "abc-123" {
		t.Fatalf("expected inbound correlation id to be honored, got %q", seen)
	}
}

func TestRecovery_ConvertsPanicTo500(t *testing.T) {
	r := gin.New()
	r.Use(Recovery(zap.NewNop()))
	r.GET("/boom", func(c *gin.Context) { panic("kaboom") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500 after panic, got %d", w.Code)
	}
}
