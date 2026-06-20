package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func newDispatcher(repo *fakeRepo, idem *fakeIdem, pub *fakePublisher) *Dispatcher {
	d := NewDispatcher(repo, idem, pub, zap.NewNop())
	d.now = func() time.Time { return time.Unix(1000, 0).UTC() }
	seq := 0
	d.newID = func() string { seq++; return "id" + string(rune('0'+seq)) }
	return d
}

func TestEnqueue_PublishesScheduleWhenEnabled(t *testing.T) {
	repo, idem, pub := newFakeRepo(), newFakeIdem(), &fakePublisher{}
	d := newDispatcher(repo, idem, pub)

	err := d.Enqueue(context.Background(), EnqueueInput{
		UserID: "u1", Type: domain.TypeDailyReminder, TemplateCode: "T", IdempotencyKey: "k1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msgs := pub.onTopic(domain.TopicSchedule)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 schedule message, got %d", len(msgs))
	}
	sm := decode[domain.ScheduleMessage](msgs[0].Value)
	if sm.StudentID != "u1" || sm.NotificationType != domain.TypeDailyReminder || sm.IdempotencyKey != "k1" {
		t.Fatalf("schedule message wrong: %+v", sm)
	}
}

func TestEnqueue_SkipsAndLogsWhenPreferenceDisabled(t *testing.T) {
	repo, idem, pub := newFakeRepo(), newFakeIdem(), &fakePublisher{}
	_ = repo.UpsertPreference(context.Background(), &domain.NotificationPreference{
		UserID: "u1", Type: domain.TypeDailyReminder, Enabled: false,
	})
	d := newDispatcher(repo, idem, pub)

	err := d.Enqueue(context.Background(), EnqueueInput{UserID: "u1", Type: domain.TypeDailyReminder, TemplateCode: "T", IdempotencyKey: "k1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pub.onTopic(domain.TopicSchedule)) != 0 {
		t.Fatalf("disabled preference must not enqueue")
	}
	if repo.logsByStatus(domain.StatusSkipped) != 1 {
		t.Fatalf("expected one SKIPPED log")
	}
}

func TestEnqueue_DeduplicatesByIdempotencyKey(t *testing.T) {
	repo, idem, pub := newFakeRepo(), newFakeIdem(), &fakePublisher{}
	d := newDispatcher(repo, idem, pub)
	in := EnqueueInput{UserID: "u1", Type: domain.TypeDailyReminder, TemplateCode: "T", IdempotencyKey: "same"}

	_ = d.Enqueue(context.Background(), in)
	_ = d.Enqueue(context.Background(), in)

	if got := len(pub.onTopic(domain.TopicSchedule)); got != 1 {
		t.Fatalf("expected exactly 1 enqueue for duplicate key, got %d", got)
	}
}

func TestEnqueue_RejectsInvalidType(t *testing.T) {
	d := newDispatcher(newFakeRepo(), newFakeIdem(), &fakePublisher{})
	err := d.Enqueue(context.Background(), EnqueueInput{UserID: "u1", Type: domain.NotificationType("BOGUS")})
	if !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
