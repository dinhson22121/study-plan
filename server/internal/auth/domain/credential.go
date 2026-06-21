package domain

import (
	"net/mail"
	"strings"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

type Role string

const (
	RoleStudent Role = "STUDENT"
	RoleAdmin   Role = "ADMIN"
)

func (r Role) Valid() bool { return r == RoleStudent || r == RoleAdmin }

type UserCredential struct {
	UserID       string
	Email        string
	PasswordHash string
	Role         Role
}

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

func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return domain.ErrValidation.WithMessage("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.ErrValidation.WithMessage("invalid email format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrValidation.WithMessage("password must be at least 8 characters")
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeEmail(email string) string { return normalizeEmail(email) }
