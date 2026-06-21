package notifhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/application"
	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

type memRepo struct {
	domain.Repository
	tokens map[string]string
	prefs  map[string]bool
}

func newMemRepo() *memRepo { return &memRepo{tokens: map[string]string{}, prefs: map[string]bool{}} }

func (r *memRepo) UpsertDeviceToken(_ context.Context, dt *domain.DeviceToken) error {
	r.tokens[dt.UserID] = dt.Token
	return nil
}
func (r *memRepo) ListPreferences(_ context.Context, _ string) ([]domain.NotificationPreference, error) {
	return nil, nil
}
func (r *memRepo) UpsertPreference(_ context.Context, p *domain.NotificationPreference) error {
	r.prefs[string(p.Type)] = p.Enabled
	return nil
}
func (r *memRepo) ListActiveUserIDs(_ context.Context) ([]string, error) { return nil, nil }

type fakeIdem struct{}

func (fakeIdem) CheckAndSet(context.Context, string, time.Duration) (bool, error) { return true, nil }

type fakePub struct{}

func (fakePub) Publish(context.Context, string, []byte, []byte) error { return nil }

func validatorFor(role string) middleware.TokenValidator {
	return func(_ context.Context, token string) (*middleware.Claims, error) {
		if len(token) > 6 && token[:6] == "valid-" {
			return &middleware.Claims{UserID: token[6:], Role: role}, nil
		}
		return nil, shared.ErrUnauthorized
	}
}

func newRouter(repo domain.Repository, role string) *gin.Engine {
	disp := application.NewDispatcher(repo, fakeIdem{}, fakePub{}, zap.NewNop())
	mgr := application.NewManager(repo, disp)
	r := gin.New()
	NewHandler(mgr, validatorFor(role)).Routes(r.Group("/api/v1"))
	return r
}

func do(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rd)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDeviceRoutes_RequireAuth(t *testing.T) {
	r := newRouter(newMemRepo(), middleware.RoleStudent)
	w := do(r, http.MethodPost, "/api/v1/devices/token", "", gin.H{"token": "x", "platform": "ios"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestRegisterToken_Success(t *testing.T) {
	repo := newMemRepo()
	r := newRouter(repo, middleware.RoleStudent)
	w := do(r, http.MethodPost, "/api/v1/devices/token", "valid-u1", gin.H{"token": "tok", "platform": "android"})
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
	if repo.tokens["u1"] != "tok" {
		t.Fatalf("token not stored")
	}
}

func TestRegisterToken_RejectsBadPlatform(t *testing.T) {
	r := newRouter(newMemRepo(), middleware.RoleStudent)
	w := do(r, http.MethodPost, "/api/v1/devices/token", "valid-u1", gin.H{"token": "tok", "platform": "symbian"})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", w.Code)
	}
}

func TestListPreferences_ReturnsDefaults(t *testing.T) {
	r := newRouter(newMemRepo(), middleware.RoleStudent)
	w := do(r, http.MethodGet, "/api/v1/notifications/preferences", "valid-u1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var env struct {
		Data []domain.NotificationPreference `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if len(env.Data) != len(domain.AllTypes()) {
		t.Fatalf("expected default prefs, got %d", len(env.Data))
	}
}

func TestSetPreference_InvalidType(t *testing.T) {
	r := newRouter(newMemRepo(), middleware.RoleStudent)
	enabled := false
	w := do(r, http.MethodPut, "/api/v1/notifications/preferences/BOGUS", "valid-u1", gin.H{"enabled": enabled})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422 for invalid type, got %d", w.Code)
	}
}

func TestBroadcast_ForbiddenForStudent(t *testing.T) {
	r := newRouter(newMemRepo(), middleware.RoleStudent)
	w := do(r, http.MethodPost, "/api/v1/admin/notifications/broadcast", "valid-u1",
		gin.H{"type": "ADMIN_BROADCAST", "template_code": "ADMIN_BROADCAST_V1"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403 for student, got %d", w.Code)
	}
}

func TestBroadcast_AllowedForAdmin(t *testing.T) {
	repo := newMemRepo()
	r := newRouter(repo, middleware.RoleAdmin)
	w := do(r, http.MethodPost, "/api/v1/admin/notifications/broadcast", "valid-admin",
		gin.H{"type": "ADMIN_BROADCAST", "template_code": "ADMIN_BROADCAST_V1"})
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 for admin, got %d: %s", w.Code, w.Body.String())
	}
}
