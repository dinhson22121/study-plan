package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newRouter(m *Metrics) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(m.Middleware())
	r.GET("/api/v1/items/:id", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/metrics", gin.WrapH(m.Handler()))
	return r
}

func TestMiddleware_CountsRequestsByRoutePattern(t *testing.T) {
	m := New()
	r := newRouter(m)

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/items/42", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	}

	got := testutil.ToFloat64(m.reqTotal.WithLabelValues("GET", "/api/v1/items/:id", "200"))
	if got != 3 {
		t.Fatalf("expected 3 counted requests on the route pattern, got %v", got)
	}
}

func TestRecordFailure_Increments(t *testing.T) {
	m := New()
	m.RecordFailure("content", "parse")
	m.RecordFailure("content", "parse")
	if got := testutil.ToFloat64(m.failures.WithLabelValues("content", "parse")); got != 2 {
		t.Fatalf("expected 2 failures, got %v", got)
	}
}

func TestHandler_ExposesMetrics(t *testing.T) {
	m := New()
	r := newRouter(m)
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/items/7", nil))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 from /metrics, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "edu_http_requests_total") {
		t.Fatalf("metrics output missing edu_http_requests_total:\n%s", w.Body.String())
	}
}
