package notification

import (
	"context"

	"go.uber.org/zap"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// preferenceSeeder creates default (all-enabled) notification preferences when a
// user registers, so the preference gate has explicit rows to consult.
type preferenceSeeder struct {
	repo domain.Repository
	log  *zap.Logger
}

func newPreferenceSeeder(repo domain.Repository, log *zap.Logger) *preferenceSeeder {
	return &preferenceSeeder{repo: repo, log: log}
}

// handle reacts to the auth UserRegisteredEvent.
func (s *preferenceSeeder) handle(ctx context.Context, evt shared.DomainEvent) error {
	e, ok := evt.(authdomain.UserRegisteredEvent)
	if !ok {
		return shared.ErrInternal.WithMessage("unexpected event type for preference seeder")
	}
	for _, p := range domain.DefaultPreferences(e.UserID) {
		pref := p // capture
		if err := s.repo.UpsertPreference(ctx, &pref); err != nil {
			return err
		}
	}
	return nil
}

// EventUserRegistered re-exports the event name the seeder subscribes to.
const EventUserRegistered = authdomain.EventUserRegistered
