package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/progress/domain"
	quizdomain "github.com/son-ngo/edu-app/internal/quiz/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	progress     map[string]*domain.TopicProgress
	streaks      map[string]*domain.Streak
	achievements map[string]bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		progress: map[string]*domain.TopicProgress{}, streaks: map[string]*domain.Streak{},
		achievements: map[string]bool{},
	}
}

func key(a, b string) string { return a + "|" + b }

func (r *fakeRepo) GetTopicProgress(_ context.Context, userID, topicID string) (*domain.TopicProgress, error) {
	if p, ok := r.progress[key(userID, topicID)]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) UpsertTopicProgress(_ context.Context, p *domain.TopicProgress) error {
	cp := *p
	r.progress[key(p.UserID, p.TopicID)] = &cp
	return nil
}
func (r *fakeRepo) ListProgressByUser(_ context.Context, userID string) ([]domain.TopicProgress, error) {
	var out []domain.TopicProgress
	for _, p := range r.progress {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (r *fakeRepo) GetStreak(_ context.Context, userID string) (*domain.Streak, error) {
	if s, ok := r.streaks[userID]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) UpsertStreak(_ context.Context, s *domain.Streak) error {
	cp := *s
	r.streaks[s.UserID] = &cp
	return nil
}
func (r *fakeRepo) HasAchievement(_ context.Context, userID string, t domain.AchievementType, ref string) (bool, error) {
	return r.achievements[userID+":"+string(t)+":"+ref], nil
}
func (r *fakeRepo) SaveAchievement(_ context.Context, a *domain.Achievement) error {
	r.achievements[a.UserID+":"+string(a.Type)+":"+a.Ref] = true
	return nil
}

type fakeTitles struct{}

func (fakeTitles) Title(context.Context, string) (string, error) { return "Logarit", nil }

type fakeNotifier struct{ pushes []map[string]string }

func (n *fakeNotifier) EnqueueReminder(_ context.Context, userID, _, _ string, vars map[string]string, _ string) error {
	m := map[string]string{"user": userID}
	for k, v := range vars {
		m[k] = v
	}
	n.pushes = append(n.pushes, m)
	return nil
}

func newService(repo *fakeRepo, notifier *fakeNotifier, now time.Time) *Service {
	s := NewService(repo, fakeTitles{}, notifier, zap.NewNop())
	s.now = func() time.Time { return now }
	return s
}

func completed(userID, topicID string, score float64) quizdomain.QuizCompletedEvent {
	return quizdomain.NewQuizCompletedEvent(userID, topicID, score, score >= 80, time.Unix(0, 0))
}

func TestHandleQuizCompleted_UpdatesMasteryAndStreak(t *testing.T) {
	repo, notifier := newFakeRepo(), &fakeNotifier{}
	svc := newService(repo, notifier, time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC))

	if err := svc.HandleQuizCompleted(context.Background(), completed("u1", "t1", 90)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := repo.progress[key("u1", "t1")]
	if p == nil || p.Status != domain.StatusCompleted || p.BestScore != 90 {
		t.Fatalf("progress not updated: %+v", p)
	}
	if repo.streaks["u1"].CurrentStreak != 1 {
		t.Fatalf("streak not started")
	}

	if len(notifier.pushes) != 1 || notifier.pushes[0]["topic"] != "Logarit" {
		t.Fatalf("expected 1 topic-completed push, got %+v", notifier.pushes)
	}
}

func TestHandleQuizCompleted_PerfectScoreTwoAchievements(t *testing.T) {
	repo, notifier := newFakeRepo(), &fakeNotifier{}
	svc := newService(repo, notifier, time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC))

	_ = svc.HandleQuizCompleted(context.Background(), completed("u1", "t1", 100))

	if len(notifier.pushes) != 2 {
		t.Fatalf("expected 2 pushes for perfect first completion, got %d", len(notifier.pushes))
	}
}

func TestHandleQuizCompleted_AchievementsAwardedOnce(t *testing.T) {
	repo, notifier := newFakeRepo(), &fakeNotifier{}
	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	svc := newService(repo, notifier, now)

	_ = svc.HandleQuizCompleted(context.Background(), completed("u1", "t1", 100))
	pushesAfterFirst := len(notifier.pushes)

	_ = svc.HandleQuizCompleted(context.Background(), completed("u1", "t1", 100))
	if len(notifier.pushes) != pushesAfterFirst {
		t.Fatalf("achievements should not be awarded twice; pushes %d -> %d", pushesAfterFirst, len(notifier.pushes))
	}
}

func TestHandleQuizCompleted_IgnoresWrongEvent(t *testing.T) {
	svc := newService(newFakeRepo(), &fakeNotifier{}, time.Unix(0, 0))
	if err := svc.HandleQuizCompleted(context.Background(), wrongEvent{}); err != nil {
		t.Fatalf("should ignore wrong event, got %v", err)
	}
}

func TestGetOverview(t *testing.T) {
	repo, notifier := newFakeRepo(), &fakeNotifier{}
	svc := newService(repo, notifier, time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC))
	_ = svc.HandleQuizCompleted(context.Background(), completed("u1", "t1", 90))
	_ = svc.HandleQuizCompleted(context.Background(), completed("u1", "t2", 50))

	ov, err := svc.GetOverview(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ov.TopicsTotal != 2 || ov.TopicsDone != 1 || ov.CurrentStreak != 1 {
		t.Fatalf("overview wrong: %+v", ov)
	}
}

type wrongEvent struct{}

func (wrongEvent) EventName() string     { return "other" }
func (wrongEvent) OccurredAt() time.Time { return time.Unix(0, 0) }
func (wrongEvent) AggregateID() string   { return "x" }
