// Package infrastructure provides the concrete adapters for the auth ports:
// bcrypt hashing, JWT issuance/validation, Postgres credential storage, and
// Redis-backed refresh-token tracking.
package infrastructure

import (
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements authdomain.Hasher using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher returns a hasher at the given cost; cost <= 0 uses bcrypt's
// default cost.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash returns the bcrypt hash of password.
func (h *BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", domain.ErrInternal.WithCause(err)
	}
	return string(b), nil
}

// Compare returns nil when password matches hash, ErrUnauthorized otherwise.
func (h *BcryptHasher) Compare(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return domain.ErrUnauthorized.WithCause(err)
	}
	return nil
}
