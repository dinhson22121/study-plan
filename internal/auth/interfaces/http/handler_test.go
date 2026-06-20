package authhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/auth/application"
	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/auth/infrastructure"
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
)

func init() { gin.SetMode(gin.TestMode) }

// memRepo is an in-memory CredentialRepository for handler tests.
type memRepo struct {
	byEmail, byID map[string]*authdomain.UserCredential
}

func newMemRepo() *memRepo {
	return &memRepo{byEmail: map[string]*authdomain.UserCredential{}, byID: map[string]*authdomain.UserCredential{}}
}
func (r *memRepo) Create(_ context.Context, c *authdomain.UserCredential) error {
	if _, ok := r.byEmail[c.Email]; ok {
		return domain.ErrConflict
	}
	cp := *c
	r.byEmail[c.Email], r.byID[c.UserID] = &cp, &cp
	return nil
}
func (r *memRepo) FindByEmail(_ context.Context, e string) (*authdomain.UserCredential, error) {
	if c, ok := r.byEmail[e]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}
func (r *memRepo) FindByUserID(_ context.Context, id string) (*authdomain.UserCredential, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}

type memStore struct{ m map[string]bool }

func newMemStore() *memStore                                            { return &memStore{m: map[string]bool{}} }
func (s *memStore) Save(_ context.Context, u, j string) error           { s.m[u+":"+j] = true; return nil }
func (s *memStore) Exists(_ context.Context, u, j string) (bool, error) { return s.m[u+":"+j], nil }
func (s *memStore) Delete(_ context.Context, u, j string) error         { delete(s.m, u+":"+j); return nil }

func newTestRouter() *gin.Engine {
	svc := application.NewService(
		newMemRepo(),
		infrastructure.NewBcryptHasher(4),
		infrastructure.NewJWTService(infrastructure.JWTConfig{
			Secret: []byte("test"), AccessTTL: time.Hour, RefreshTTL: time.Hour, Issuer: "t",
		}),
		newMemStore(),
		eventbus.New(),
	)
	r := gin.New()
	NewHandler(svc, svc.ValidateAccessToken).Routes(r.Group("/api/v1"))
	return r
}

func doJSON(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegisterEndpoint_CreatesAccount(t *testing.T) {
	r := newTestRouter()
	w := doJSON(r, http.MethodPost, "/api/v1/auth/register", gin.H{"email": "a@b.com", "password": "password1"})
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if !env.Success || env.Data.AccessToken == "" {
		t.Fatalf("expected tokens in response, got %s", w.Body.String())
	}
}

func TestRegisterEndpoint_ValidationError(t *testing.T) {
	r := newTestRouter()
	w := doJSON(r, http.MethodPost, "/api/v1/auth/register", gin.H{"email": "bad", "password": "x"})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLoginEndpoint_FlowAndWrongPassword(t *testing.T) {
	r := newTestRouter()
	_ = doJSON(r, http.MethodPost, "/api/v1/auth/register", gin.H{"email": "a@b.com", "password": "password1"})

	ok := doJSON(r, http.MethodPost, "/api/v1/auth/login", gin.H{"email": "a@b.com", "password": "password1"})
	if ok.Code != http.StatusOK {
		t.Fatalf("want 200 on valid login, got %d", ok.Code)
	}
	bad := doJSON(r, http.MethodPost, "/api/v1/auth/login", gin.H{"email": "a@b.com", "password": "wrongpass"})
	if bad.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 on wrong password, got %d", bad.Code)
	}
}

func TestLogoutEndpoint_RequiresAuth(t *testing.T) {
	r := newTestRouter()
	// No Authorization header -> logout must be rejected even with a refresh token.
	w := doJSON(r, http.MethodPost, "/api/v1/auth/logout", gin.H{"refresh_token": "whatever"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 for unauthenticated logout, got %d", w.Code)
	}
}

func TestRegisterEndpoint_DuplicateConflict(t *testing.T) {
	r := newTestRouter()
	_ = doJSON(r, http.MethodPost, "/api/v1/auth/register", gin.H{"email": "a@b.com", "password": "password1"})
	dup := doJSON(r, http.MethodPost, "/api/v1/auth/register", gin.H{"email": "a@b.com", "password": "password1"})
	if dup.Code != http.StatusConflict {
		t.Fatalf("want 409 on duplicate, got %d", dup.Code)
	}
}
