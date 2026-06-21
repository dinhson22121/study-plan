package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/shared/audit"
)

// fakeRecorder captures entries passed to Record for assertions.
type fakeRecorder struct {
	mu      sync.Mutex
	entries []audit.Entry
	err     error
}

func (f *fakeRecorder) Record(_ context.Context, e audit.Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, e)
	return f.err
}

func (f *fakeRecorder) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.entries)
}

// auditRouter builds a router whose /thing route fakes per-route Auth (setting
// role/user the way middleware.Auth would) and returns the configured status.
// AuditAdmin is registered at the group level, mirroring bootstrap wiring.
func auditRouter(rec audit.Recorder, role, userID string, status int) *gin.Engine {
	r := gin.New()
	r.Use(Logger(zap.NewNop()))
	g := r.Group("/api")
	g.Use(AuditAdmin(rec, zap.NewNop()))

	handler := func(c *gin.Context) {
		c.Set(ctxUserID, userID)
		c.Set(ctxRole, role)
		c.Status(status)
	}
	g.POST("/thing", handler)
	g.PUT("/thing", handler)
	g.DELETE("/thing", handler)
	g.GET("/thing", handler)
	return r
}

func doRequest(t *testing.T, r *gin.Engine, method, path string) {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
}

func TestAuditAdmin_RecordsOnAdminMutating2xx(t *testing.T) {
	rec := &fakeRecorder{}
	r := auditRouter(rec, RoleAdmin, "admin1", http.StatusCreated)

	doRequest(t, r, http.MethodPost, "/api/thing")

	if rec.count() != 1 {
		t.Fatalf("want 1 audit entry, got %d", rec.count())
	}
	e := rec.entries[0]
	if e.ActorUserID != "admin1" {
		t.Fatalf("actor user id = %q, want admin1", e.ActorUserID)
	}
	if e.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", e.Method)
	}
	if e.Path != "/api/thing" {
		t.Fatalf("path = %q, want /api/thing (FullPath)", e.Path)
	}
	if e.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", e.StatusCode)
	}
	if e.CorrelationID == "" {
		t.Fatalf("expected correlation id to be captured")
	}
	if e.Action() != "POST /api/thing" {
		t.Fatalf("action = %q, want \"POST /api/thing\"", e.Action())
	}
}

func TestAuditAdmin_RecordsOnPutAndDelete(t *testing.T) {
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		rec := &fakeRecorder{}
		r := auditRouter(rec, RoleAdmin, "admin1", http.StatusOK)
		doRequest(t, r, method, "/api/thing")
		if rec.count() != 1 {
			t.Fatalf("%s: want 1 audit entry, got %d", method, rec.count())
		}
	}
}

func TestAuditAdmin_SkipsGet(t *testing.T) {
	rec := &fakeRecorder{}
	r := auditRouter(rec, RoleAdmin, "admin1", http.StatusOK)

	doRequest(t, r, http.MethodGet, "/api/thing")

	if rec.count() != 0 {
		t.Fatalf("want 0 audit entries for GET, got %d", rec.count())
	}
}

func TestAuditAdmin_SkipsNonAdmin(t *testing.T) {
	rec := &fakeRecorder{}
	r := auditRouter(rec, RoleStudent, "student1", http.StatusCreated)

	doRequest(t, r, http.MethodPost, "/api/thing")

	if rec.count() != 0 {
		t.Fatalf("want 0 audit entries for non-admin, got %d", rec.count())
	}
}

func TestAuditAdmin_SkipsMissingRole(t *testing.T) {
	rec := &fakeRecorder{}
	r := auditRouter(rec, "", "", http.StatusCreated)

	doRequest(t, r, http.MethodPost, "/api/thing")

	if rec.count() != 0 {
		t.Fatalf("want 0 audit entries when role unset, got %d", rec.count())
	}
}

func TestAuditAdmin_SkipsClientError(t *testing.T) {
	rec := &fakeRecorder{}
	r := auditRouter(rec, RoleAdmin, "admin1", http.StatusBadRequest)

	doRequest(t, r, http.MethodPost, "/api/thing")

	if rec.count() != 0 {
		t.Fatalf("want 0 audit entries for 4xx, got %d", rec.count())
	}
}

func TestAuditAdmin_SkipsServerError(t *testing.T) {
	rec := &fakeRecorder{}
	r := auditRouter(rec, RoleAdmin, "admin1", http.StatusInternalServerError)

	doRequest(t, r, http.MethodPost, "/api/thing")

	if rec.count() != 0 {
		t.Fatalf("want 0 audit entries for 5xx, got %d", rec.count())
	}
}

func TestAuditAdmin_RecorderErrorDoesNotFailRequest(t *testing.T) {
	rec := &fakeRecorder{err: context.DeadlineExceeded}
	r := auditRouter(rec, RoleAdmin, "admin1", http.StatusCreated)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/thing", nil))

	if w.Code != http.StatusCreated {
		t.Fatalf("recorder error must not change response: got %d, want 201", w.Code)
	}
	if rec.count() != 1 {
		t.Fatalf("expected Record to have been attempted once, got %d", rec.count())
	}
}
