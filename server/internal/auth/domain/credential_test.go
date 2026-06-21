package domain

import (
	"errors"
	"strings"
	"testing"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestNewUserCredential_ValidatesAndNormalizes(t *testing.T) {
	c, err := NewUserCredential("u1", "  Alice@Example.COM ", "hash", RoleStudent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Email != "alice@example.com" {
		t.Fatalf("email not normalized: %q", c.Email)
	}
}

func TestNewUserCredential_RejectsBadInput(t *testing.T) {
	cases := []struct {
		name, email, hash string
		role              Role
	}{
		{"empty email", "", "h", RoleStudent},
		{"bad email", "not-an-email", "h", RoleStudent},
		{"empty hash", "a@b.com", "", RoleStudent},
		{"bad role", "a@b.com", "h", Role("KING")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUserCredential("u1", tc.email, tc.hash, tc.role)
			if !errors.Is(err, shared.ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	rejected := []struct{ name, pw string }{
		{"too short", "abc12"},
		{"nine chars", "abcdefg12"},
		{"no digit", "abcdefghij"},
		{"no letter", "1234567890"},
		{"too long", strings.Repeat("a1", 37)},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidatePassword(tc.pw); !errors.Is(err, shared.ErrValidation) {
				t.Fatalf("expected validation error for %q, got %v", tc.pw, err)
			}
		})
	}

	accepted := []string{"abcdefgh12", "Str0ngPass!", strings.Repeat("a1", 36)}
	for _, pw := range accepted {
		if err := ValidatePassword(pw); err != nil {
			t.Fatalf("unexpected error for valid password %q: %v", pw, err)
		}
	}
}
