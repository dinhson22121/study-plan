// Package domain defines the auth bounded context: credentials, tokens, the
// ports auth depends on, and the events it publishes.
package domain

import (
	"net/mail"
	"strings"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// Role enumerates account roles. Auth is the source of truth for a user's role.
type Role string

const (
	RoleStudent Role = "STUDENT"
	RoleAdmin   Role = "ADMIN"
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool { return r == RoleStudent || r == RoleAdmin }

// UserCredential is the authentication aggregate: the identity (UserID), login
// email, hashed password, and role. The plaintext password never lives here.
type UserCredential struct {
	UserID       string
	Email        string
	PasswordHash string
	Role         Role
}

// NewUserCredential constructs a credential after validating email and role. The
// password is already hashed by the application layer (via the Hasher port).
func NewUserCredential(userID, email, passwordHash string, role Role) (*UserCredential, error) {
	email = normalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if !role.Valid() {
		return nil, domain.ErrValidation.WithMessage("invalid role")
	}
	if passwordHash == "" {
		return nil, domain.ErrValidation.WithMessage("password hash required")
	}
	return &UserCredential{UserID: userID, Email: email, PasswordHash: passwordHash, Role: role}, nil
}

// ValidateEmail checks that an address is syntactically valid and non-empty.
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return domain.ErrValidation.WithMessage("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.ErrValidation.WithMessage("invalid email format")
	}
	return nil
}

// ValidatePassword enforces the minimum password policy at the boundary.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrValidation.WithMessage("password must be at least 8 characters")
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeEmail exposes canonical email formatting for lookups.
func NormalizeEmail(email string) string { return normalizeEmail(email) }
