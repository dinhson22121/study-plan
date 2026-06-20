package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

func init() { gin.SetMode(gin.TestMode) }

func run(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	r := gin.New()
	r.GET("/", handler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	return w
}

func TestOK_WrapsDataInSuccessEnvelope(t *testing.T) {
	w := run(func(c *gin.Context) { OK(c, gin.H{"x": 1}) })
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var env Envelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if !env.Success {
		t.Fatalf("expected success=true")
	}
}

func TestCreated_Returns201(t *testing.T) {
	w := run(func(c *gin.Context) { Created(c, gin.H{}) })
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
}

func TestList_IncludesMeta(t *testing.T) {
	w := run(func(c *gin.Context) { List(c, []int{1, 2}, Meta{Total: 2, Page: 1, Limit: 10}) })
	var env Envelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if env.Meta == nil || env.Meta.Total != 2 {
		t.Fatalf("expected meta with total=2, got %+v", env.Meta)
	}
}

func TestFail_MapsErrorToStatusAndEnvelope(t *testing.T) {
	w := run(func(c *gin.Context) { Fail(c, domain.ErrNotFound) })
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
	var env Envelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if env.Success || env.Error == nil || env.Error.Code != domain.ErrNotFound.Code {
		t.Fatalf("expected error envelope with NOT_FOUND, got %+v", env)
	}
}

func TestFail_DuplicateIsIdempotent200(t *testing.T) {
	w := run(func(c *gin.Context) { Fail(c, domain.ErrDuplicateMessage) })
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 for idempotent duplicate, got %d", w.Code)
	}
	var env Envelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if !env.Success {
		t.Fatalf("expected success=true for idempotent replay")
	}
}
