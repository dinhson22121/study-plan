package domain

import (
	"context"
	"strings"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type User struct {
	ID          string
	Email       string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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

type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, u *User) error
}
