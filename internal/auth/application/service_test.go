package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
)

// --- in-memory fakes for the auth ports ---

type fakeRepo struct {
	byEmail  map[string]*authdomain.UserCredential
	byUserID map[string]*authdomain.UserCredential
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byEmail: map[string]*authdomain.UserCredential{}, byUserID: map[string]*authdomain.UserCredential{}}
}

func (r *fakeRepo) Create(_ context.Context, c *authdomain.UserCredential) error {
	if _, ok := r.byEmail[c.Email]; ok {
		return domain.ErrConflict
	}
	cp := *c
	r.byEmail[c.Email] = &cp
	r.byUserID[c.UserID] = &cp
	return nil
}
func (r *fakeRepo) FindByEmail(_ context.Context, email string) (*authdomain.UserCredential, error) {
	if c, ok := r.byEmail[email]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}
func (r *fakeRepo) FindByUserID(_ context.Context, id string) (*authdomain.UserCredential, error) {
	if c, ok := r.byUserID[id]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}

type fakeHasher struct{}

func (fakeHasher) Hash(pw string) (string, error) { return "hashed:" + pw, nil }
func (fakeHasher) Compare(hash, pw string) error {
	if hash == "hashed:"+pw {
		return nil
	}
	return domain.ErrUnauthorized
}

type fakeTokens struct{ jtiSeq int }

func (t *fakeTokens) IssueAccess(userID string, role authdomain.Role) (string, int64, error) {
	return "A|" + userID + "|" + string(role), time.Unix(1000, 0).Unix(), nil
}
func (t *fakeTokens) IssueRefresh(userID string) (string, string, error) {
	t.jtiSeq++
	jti := "jti" + strconv.Itoa(t.jtiSeq)
	return "R|" + userID + "|" + jti, jti, nil
}
func (t *fakeTokens) ParseAccess(token string) (*authdomain.Claims, error) {
	p := strings.Split(token, "|")
	if len(p) != 3 || p[0] != "A" {
		return nil, domain.ErrUnauthorized
	}
	return &authdomain.Claims{UserID: p[1], Role: authdomain.Role(p[2])}, nil
}
func (t *fakeTokens) ParseRefresh(token string) (*authdomain.RefreshClaims, error) {
	p := strings.Split(token, "|")
	if len(p) != 3 || p[0] != "R" {
		return nil, domain.ErrUnauthorized
	}
	return &authdomain.RefreshClaims{UserID: p[1], ID: p[2]}, nil
}

type fakeRefreshStore struct{ active map[string]bool }

func newFakeRefreshStore() *fakeRefreshStore { return &fakeRefreshStore{active: map[string]bool{}} }
func key(u, j string) string                 { return u + ":" + j }
func (s *fakeRefreshStore) Save(_ context.Context, u, j string) error {
	s.active[key(u, j)] = true
	return nil
}
func (s *fakeRefreshStore) Exists(_ context.Context, u, j string) (bool, error) {
	return s.active[key(u, j)], nil
}
func (s *fakeRefreshStore) Delete(_ context.Context, u, j string) error {
	delete(s.active, key(u, j))
	return nil
}

func newTestService(t *testing.T) (*Service, *fakeRepo, *fakeRefreshStore, *eventbus.Bus) {
	t.Helper()
	repo := newFakeRepo()
	store := newFakeRefreshStore()
	bus := eventbus.New()
	svc := NewService(repo, fakeHasher{}, &fakeTokens{}, store, bus,
		WithClock(func() time.Time { return time.Unix(500, 0) }))
	return svc, repo, store, bus
}

// --- tests ---

func TestRegister_CreatesCredentialPublishesEventAndReturnsTokens(t *testing.T) {
	svc, repo, store, bus := newTestService(t)
	var gotEvent *authdomain.UserRegisteredEvent
	bus.Subscribe(authdomain.EventUserRegistered, func(_ context.Context, e domain.DomainEvent) error {
		ev := e.(authdomain.UserRegisteredEvent)
		gotEvent = &ev
		return nil
	})

	pair, err := svc.Register(context.Background(), RegisterInput{Email: "A@b.com", Password: "password1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected tokens, got %+v", pair)
	}
	if _, ok := repo.byEmail["a@b.com"]; !ok {
		t.Fatalf("credential not stored under normalized email")
	}
	if gotEvent == nil || gotEvent.Email != "a@b.com" || gotEvent.Role != authdomain.RoleStudent {
		t.Fatalf("UserRegisteredEvent not published correctly: %+v", gotEvent)
	}
	if len(store.active) != 1 {
		t.Fatalf("expected one active refresh token, got %d", len(store.active))
	}
}

func TestRegister_RejectsDuplicateEmail(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})
	_, err := svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestRegister_RejectsWeakPassword(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	_, err := svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "short"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRegister_FailsWhenEventHandlerFails(t *testing.T) {
	svc, _, _, bus := newTestService(t)
	bus.Subscribe(authdomain.EventUserRegistered, func(context.Context, domain.DomainEvent) error {
		return errors.New("profile creation failed")
	})
	_, err := svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})
	if !errors.Is(err, domain.ErrInternal) {
		t.Fatalf("expected internal error when handler fails, got %v", err)
	}
}

func TestLogin_SucceedsWithCorrectPassword(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})
	pair, err := svc.Login(context.Background(), LoginInput{Email: "A@b.com", Password: "password1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.AccessToken == "" {
		t.Fatalf("expected access token")
	}
}

func TestLogin_FailsWithWrongPasswordOrUnknownEmail(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	_, _ = svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})

	if _, err := svc.Login(context.Background(), LoginInput{Email: "a@b.com", Password: "wrongpass"}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for wrong password, got %v", err)
	}
	if _, err := svc.Login(context.Background(), LoginInput{Email: "ghost@b.com", Password: "password1"}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for unknown email, got %v", err)
	}
}

func TestRefresh_RotatesAndRevokesOldToken(t *testing.T) {
	svc, _, store, _ := newTestService(t)
	pair, _ := svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})

	newPair, err := svc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPair.RefreshToken == pair.RefreshToken {
		t.Fatalf("refresh token should rotate")
	}
	// Old token must now be revoked: reusing it fails.
	if _, err := svc.Refresh(context.Background(), pair.RefreshToken); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected reused old refresh token to be rejected, got %v", err)
	}
	if len(store.active) != 1 {
		t.Fatalf("expected exactly one active refresh token after rotation, got %d", len(store.active))
	}
}

func TestLogout_RevokesRefreshToken(t *testing.T) {
	svc, _, store, _ := newTestService(t)
	pair, _ := svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})

	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.active) != 0 {
		t.Fatalf("expected no active tokens after logout, got %d", len(store.active))
	}
	if _, err := svc.Refresh(context.Background(), pair.RefreshToken); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected refresh after logout to fail, got %v", err)
	}
}

func TestValidateAccessToken(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	pair, _ := svc.Register(context.Background(), RegisterInput{Email: "a@b.com", Password: "password1"})

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Role != string(authdomain.RoleStudent) {
		t.Fatalf("expected STUDENT role, got %s", claims.Role)
	}
	if _, err := svc.ValidateAccessToken("garbage"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for garbage token, got %v", err)
	}
}
