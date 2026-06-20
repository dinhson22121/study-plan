package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type seederRepo struct {
	domain.Repository
	upserts []domain.NotificationPreference
}

func (r *seederRepo) UpsertPreference(_ context.Context, p *domain.NotificationPreference) error {
	r.upserts = append(r.upserts, *p)
	return nil
}

func TestPreferenceSeeder_SeedsAllDefaultsOnRegistration(t *testing.T) {
	repo := &seederRepo{}
	s := newPreferenceSeeder(repo, zap.NewNop())

	evt := authdomain.NewUserRegisteredEvent("u1", "a@b.com", authdomain.RoleStudent, time.Unix(0, 0))
	if err := s.handle(context.Background(), evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.upserts) != len(domain.AllTypes()) {
		t.Fatalf("expected %d preferences seeded, got %d", len(domain.AllTypes()), len(repo.upserts))
	}
	for _, p := range repo.upserts {
		if p.UserID != "u1" || !p.Enabled {
			t.Fatalf("seeded preference wrong: %+v", p)
		}
	}
}

func TestPreferenceSeeder_RejectsWrongEventType(t *testing.T) {
	s := newPreferenceSeeder(&seederRepo{}, zap.NewNop())
	if err := s.handle(context.Background(), wrongEvent{}); !errors.Is(err, shared.ErrInternal) {
		t.Fatalf("expected internal error for wrong event type, got %v", err)
	}
}

type wrongEvent struct{}

func (wrongEvent) EventName() string     { return "other" }
func (wrongEvent) OccurredAt() time.Time { return time.Unix(0, 0) }
func (wrongEvent) AggregateID() string   { return "x" }
