package userhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/internal/user/application"
	userdomain "github.com/son-ngo/edu-app/internal/user/domain"
)

func init() { gin.SetMode(gin.TestMode) }

type memRepo struct{ users map[string]*userdomain.User }

func newMemRepo() *memRepo { return &memRepo{users: map[string]*userdomain.User{}} }
func (r *memRepo) Create(_ context.Context, u *userdomain.User) error {
	cp := *u
	r.users[u.ID] = &cp
	return nil
}
func (r *memRepo) FindByID(_ context.Context, id string) (*userdomain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}
func (r *memRepo) Update(_ context.Context, u *userdomain.User) error {
	if _, ok := r.users[u.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *u
	r.users[u.ID] = &cp
	return nil
}

// validatorFor returns a TokenValidator that accepts "valid-<userID>" tokens.
func validatorFor() middleware.TokenValidator {
	return func(token string) (*middleware.Claims, error) {
		if len(token) > 6 && token[:6] == "valid-" {
			return &middleware.Claims{UserID: token[6:], Role: middleware.RoleStudent}, nil
		}
		return nil, domain.ErrUnauthorized
	}
}

func newRouter(repo *memRepo) *gin.Engine {
	svc := application.NewService(repo, application.WithClock(func() time.Time { return time.Unix(1000, 0) }))
	r := gin.New()
	NewHandler(svc, validatorFor()).Routes(r.Group("/api/v1"))
	return r
}

func TestGetMe_RequiresAuth(t *testing.T) {
	r := newRouter(newMemRepo())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 without token, got %d", w.Code)
	}
}

func TestGetMe_ReturnsProfile(t *testing.T) {
	repo := newMemRepo()
	u, _ := userdomain.NewUser("u1", "a@b.com", "Alice", time.Unix(0, 0))
	_ = repo.Create(context.Background(), u)
	r := newRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer valid-u1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateMe_ChangesDisplayName(t *testing.T) {
	repo := newMemRepo()
	u, _ := userdomain.NewUser("u1", "a@b.com", "Alice", time.Unix(0, 0))
	_ = repo.Create(context.Background(), u)
	r := newRouter(repo)

	b, _ := json.Marshal(gin.H{"display_name": "Alice B"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer valid-u1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.users["u1"].DisplayName != "Alice B" {
		t.Fatalf("display name not updated: %q", repo.users["u1"].DisplayName)
	}
}

func TestUpdateMe_ValidationError(t *testing.T) {
	repo := newMemRepo()
	u, _ := userdomain.NewUser("u1", "a@b.com", "Alice", time.Unix(0, 0))
	_ = repo.Create(context.Background(), u)
	r := newRouter(repo)

	b, _ := json.Marshal(gin.H{}) // missing display_name
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer valid-u1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", w.Code)
	}
}
