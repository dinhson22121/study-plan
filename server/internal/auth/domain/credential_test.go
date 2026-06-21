package domain

import (
	"errors"
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
	if err := ValidatePassword("short"); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for short password")
	}
	if err := ValidatePassword("longenough"); err != nil {
		t.Fatalf("unexpected error for valid password: %v", err)
	}
}
