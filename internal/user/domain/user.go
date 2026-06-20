// Package domain defines the user bounded context: the User profile aggregate
// and its repository port. Identity/credentials live in the auth context; this
// context owns profile data only.
package domain

import (
	"context"
	"strings"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// User is the profile aggregate. ID equals the identity created by auth.
type User struct {
	ID          string
	Email       string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewUser builds a profile, deriving a default display name from the email local
// part when none is supplied.
func NewUser(id, email, displayName string, now time.Time) (*User, error) {
	if id == "" {
		return nil, shared.ErrValidation.WithMessage("user id required")
	}
	if email == "" {
		return nil, shared.ErrValidation.WithMessage("email required")
	}
	if strings.TrimSpace(displayName) == "" {
		displayName = deriveDisplayName(email)
	}
	return &User{ID: id, Email: email, DisplayName: displayName, CreatedAt: now, UpdatedAt: now}, nil
}

// Rename returns a new User with an updated display name and timestamp
// (immutable update — the receiver is not mutated).
func (u User) Rename(displayName string, now time.Time) (*User, error) {
	if strings.TrimSpace(displayName) == "" {
		return nil, shared.ErrValidation.WithMessage("display name cannot be empty")
	}
	u.DisplayName = strings.TrimSpace(displayName)
	u.UpdatedAt = now
	return &u, nil
}

func deriveDisplayName(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
}

// Repository persists and retrieves user profiles.
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, u *User) error
}
