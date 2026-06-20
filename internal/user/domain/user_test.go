package domain

import (
	"errors"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestNewUser_DerivesDisplayNameFromEmail(t *testing.T) {
	now := time.Unix(1000, 0)
	u, err := NewUser("u1", "minh@example.com", "", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.DisplayName != "minh" {
		t.Fatalf("expected derived name 'minh', got %q", u.DisplayName)
	}
	if !u.CreatedAt.Equal(now) || !u.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps not set")
	}
}

func TestNewUser_KeepsProvidedDisplayName(t *testing.T) {
	u, _ := NewUser("u1", "minh@example.com", "Minh Nguyen", time.Unix(0, 0))
	if u.DisplayName != "Minh Nguyen" {
		t.Fatalf("expected provided name, got %q", u.DisplayName)
	}
}

func TestNewUser_Validation(t *testing.T) {
	if _, err := NewUser("", "a@b.com", "x", time.Unix(0, 0)); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for empty id")
	}
	if _, err := NewUser("u1", "", "x", time.Unix(0, 0)); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for empty email")
	}
}

func TestUser_RenameIsImmutable(t *testing.T) {
	orig, _ := NewUser("u1", "a@b.com", "Old", time.Unix(0, 0))
	later := time.Unix(2000, 0)

	updated, err := orig.Rename("New", later)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orig.DisplayName != "Old" {
		t.Fatalf("original was mutated: %q", orig.DisplayName)
	}
	if updated.DisplayName != "New" || !updated.UpdatedAt.Equal(later) {
		t.Fatalf("updated copy incorrect: %+v", updated)
	}
	if _, err := orig.Rename("   ", later); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for blank name")
	}
}
