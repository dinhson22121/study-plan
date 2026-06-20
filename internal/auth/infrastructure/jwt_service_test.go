package infrastructure

import (
	"errors"
	"testing"
	"time"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

func newJWT(ttl time.Duration) *JWTService {
	return NewJWTService(JWTConfig{
		Secret:     []byte("test-secret"),
		AccessTTL:  ttl,
		RefreshTTL: time.Hour,
		Issuer:     "edu-app-test",
	})
}

func TestJWT_AccessRoundTrip(t *testing.T) {
	s := newJWT(time.Hour)
	token, exp, err := s.IssueAccess("user-1", authdomain.RoleAdmin)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if exp <= time.Now().Unix() {
		t.Fatalf("expiry should be in the future")
	}
	claims, err := s.ParseAccess(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-1" || claims.Role != authdomain.RoleAdmin {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestJWT_RefreshRoundTripCarriesJTI(t *testing.T) {
	s := newJWT(time.Hour)
	token, jti, err := s.IssueRefresh("user-2")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := s.ParseRefresh(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-2" || claims.ID != jti {
		t.Fatalf("refresh claims mismatch: got %+v want jti %s", claims, jti)
	}
}

func TestJWT_RejectsExpiredToken(t *testing.T) {
	s := newJWT(-time.Minute) // already expired
	token, _, _ := s.IssueAccess("user-3", authdomain.RoleStudent)
	if _, err := s.ParseAccess(token); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for expired token, got %v", err)
	}
}

func TestJWT_RejectsTamperedToken(t *testing.T) {
	s := newJWT(time.Hour)
	token, _, _ := s.IssueAccess("user-4", authdomain.RoleStudent)
	if _, err := s.ParseAccess(token + "x"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for tampered token, got %v", err)
	}
}

func TestBcryptHasher_HashAndCompare(t *testing.T) {
	h := NewBcryptHasher(4) // low cost for fast tests
	hash, err := h.Hash("password1")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := h.Compare(hash, "password1"); err != nil {
		t.Fatalf("compare valid: %v", err)
	}
	if err := h.Compare(hash, "wrongpass"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for wrong password, got %v", err)
	}
}
