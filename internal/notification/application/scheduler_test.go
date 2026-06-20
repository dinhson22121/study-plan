package application

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
)

func newTestScheduler(repo *fakeRepo, pub *fakePublisher) *Scheduler {
	disp := NewDispatcher(repo, newFakeIdem(), pub, zap.NewNop())
	return NewScheduler(disp, repo, zap.NewNop(), "Asia/Ho_Chi_Minh")
}

func TestScheduler_DailyReminderFansOutToActiveUsers(t *testing.T) {
	repo := newFakeRepo()
	repo.activeUsers = []string{"u1", "u2"}
	pub := &fakePublisher{}
	s := newTestScheduler(repo, pub)

	s.runDailyReminder()

	msgs := pub.onTopic(domain.TopicSchedule)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 daily reminders enqueued, got %d", len(msgs))
	}
	sm := decode[domain.ScheduleMessage](msgs[0].Value)
	if sm.NotificationType != domain.TypeDailyReminder || sm.TemplateCode != TemplateDailyReminder {
		t.Fatalf("wrong schedule message: %+v", sm)
	}
}

func TestScheduler_RespectsPreferenceGate(t *testing.T) {
	repo := newFakeRepo()
	repo.activeUsers = []string{"u1", "u2"}
	_ = repo.UpsertPreference(context.Background(), &domain.NotificationPreference{
		UserID: "u1", Type: domain.TypeWeeklyQuiz, Enabled: false,
	})
	pub := &fakePublisher{}
	s := newTestScheduler(repo, pub)

	s.runWeeklyQuiz()

	if got := len(pub.onTopic(domain.TopicSchedule)); got != 1 {
		t.Fatalf("expected 1 enqueue (u1 gated off), got %d", got)
	}
}

func TestScheduler_ReengagementIsNoOp(t *testing.T) {
	repo := newFakeRepo()
	repo.activeUsers = []string{"u1"}
	pub := &fakePublisher{}
	s := newTestScheduler(repo, pub)

	s.runReengagement()

	if len(pub.messages) != 0 {
		t.Fatalf("reengagement should not enqueue yet (pending analytics)")
	}
}
