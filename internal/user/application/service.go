// Package application contains the user profile use cases and the reaction to
// the auth UserRegisteredEvent that creates a profile on signup.
package application

import (
	"context"
	"time"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	userdomain "github.com/son-ngo/edu-app/internal/user/domain"
)

// Service implements the user profile use cases.
type Service struct {
	repo userdomain.Repository
	now  func() time.Time
}

// Option customizes a Service.
type Option func(*Service)

// WithClock overrides the time source (tests).
func WithClock(now func() time.Time) Option { return func(s *Service) { s.now = now } }

// NewService builds the user service.
func NewService(repo userdomain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, o := range opts {
		o(s)
	}
	return s
}

// HandleUserRegistered creates a profile in response to the auth event. It is
// idempotent at the storage layer: a duplicate (replayed) event for an existing
// user id is treated as success.
func (s *Service) HandleUserRegistered(ctx context.Context, evt shared.DomainEvent) error {
	e, ok := evt.(authdomain.UserRegisteredEvent)
	if !ok {
		return shared.ErrInternal.WithMessage("unexpected event type for user.registered handler")
	}
	u, err := userdomain.NewUser(e.UserID, e.Email, "", s.now())
	if err != nil {
		return err
	}
	if err := s.repo.Create(ctx, u); err != nil {
		if shared.AsDomainError(err).Code == shared.ErrConflict.Code {
			return nil // already created — idempotent
		}
		return err
	}
	return nil
}

// GetProfile returns the profile for a user id.
func (s *Service) GetProfile(ctx context.Context, userID string) (*userdomain.User, error) {
	return s.repo.FindByID(ctx, userID)
}

// UpdateDisplayName renames a user and persists the change.
func (s *Service) UpdateDisplayName(ctx context.Context, userID, displayName string) (*userdomain.User, error) {
	current, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	updated, err := current.Rename(displayName, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, updated); err != nil {
		return nil, err
	}
	return updated, nil
}
