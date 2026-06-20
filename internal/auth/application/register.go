package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// RegisterInput is the registration command payload.
type RegisterInput struct {
	Email    string
	Password string
}

// Register creates a new STUDENT account: it validates input, ensures the email
// is unused, hashes the password, persists the credential, publishes
// UserRegisteredEvent (so user/notification can react), and returns a token
// pair so the client is logged in immediately.
func (s *Service) Register(ctx context.Context, in RegisterInput) (authdomain.TokenPair, error) {
	var zero authdomain.TokenPair

	email := authdomain.NormalizeEmail(in.Email)
	if err := authdomain.ValidateEmail(email); err != nil {
		return zero, err
	}
	if err := authdomain.ValidatePassword(in.Password); err != nil {
		return zero, err
	}

	if existing, err := s.repo.FindByEmail(ctx, email); err == nil && existing != nil {
		return zero, domain.ErrConflict.WithMessage("email already registered")
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return zero, err
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return zero, domain.ErrInternal.WithCause(err)
	}

	userID := uuid.NewString()
	cred, err := authdomain.NewUserCredential(userID, email, hash, authdomain.RoleStudent)
	if err != nil {
		return zero, err
	}
	if err := s.repo.Create(ctx, cred); err != nil {
		return zero, err
	}

	// Publish synchronously so the user profile and notification preferences are
	// created within the request. A handler failure fails registration; an
	// outbox/saga would make this fully atomic (see plan, future work).
	evt := authdomain.NewUserRegisteredEvent(userID, email, authdomain.RoleStudent, s.now())
	if err := s.bus.Publish(ctx, evt); err != nil {
		return zero, domain.ErrInternal.WithCause(err)
	}

	return s.issueTokenPair(ctx, userID, authdomain.RoleStudent)
}
