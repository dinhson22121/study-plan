package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/analytics/domain"
	quizdomain "github.com/son-ngo/edu-app/internal/quiz/domain"
)

type fakeActivity struct {
	appended  []string
	inactive  []string
	lastSince time.Time
}

func (f *fakeActivity) Append(_ context.Context, userID string, _ time.Time) error {
	f.appended = append(f.appended, userID)
	return nil
}
func (f *fakeActivity) InactiveUserIDs(_ context.Context, before time.Time) ([]string, error) {
	f.lastSince = before
	return f.inactive, nil
}

type fakeProgress struct{ snap domain.ProgressSnapshot }

func (f fakeProgress) Snapshot(context.Context, string) (domain.ProgressSnapshot, error) {
	return f.snap, nil
}

type fakeQuiz struct{ scores []float64 }

func (f fakeQuiz) Scores(context.Context, string) ([]float64, error) { return f.scores, nil }

func newService(act *fakeActivity, prog fakeProgress, qz fakeQuiz, now time.Time) *Service {
	s := NewService(act, prog, qz, zap.NewNop())
	s.now = func() time.Time { return now }
	return s
}

func TestDashboard_AggregatesAverage(t *testing.T) {
	prog := fakeProgress{snap: domain.ProgressSnapshot{CurrentStreak: 3, LongestStreak: 5, TopicsTotal: 4, TopicsCompleted: 2}}
	qz := fakeQuiz{scores: []float64{80, 100, 60}} // avg 80
	svc := newService(&fakeActivity{}, prog, qz, time.Unix(0, 0))

	d, err := svc.Dashboard(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.QuizAverage != 80 || d.QuizCount != 3 || d.TopicsCompleted != 2 || d.CurrentStreak != 3 {
		t.Fatalf("dashboard wrong: %+v", d)
	}
}

func TestDashboard_NoQuizzes(t *testing.T) {
	svc := newService(&fakeActivity{}, fakeProgress{}, fakeQuiz{}, time.Unix(0, 0))
	d, err := svc.Dashboard(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.QuizAverage != 0 || d.QuizCount != 0 {
		t.Fatalf("expected zero average, got %+v", d)
	}
}

func TestWeakTopics_RanksNotCompletedAsc(t *testing.T) {
	prog := fakeProgress{snap: domain.ProgressSnapshot{Topics: []domain.TopicStat{
		{TopicID: "a", Completed: true, BestScore: 90},
		{TopicID: "b", Completed: false, BestScore: 40},
		{TopicID: "c", Completed: false, BestScore: 20},
		{TopicID: "d", Completed: false, BestScore: 60},
	}}}
	svc := newService(&fakeActivity{}, prog, fakeQuiz{}, time.Unix(0, 0))

	weak, err := svc.WeakTopics(context.Background(), "u1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(weak) != 2 || weak[0].TopicID != "c" || weak[1].TopicID != "b" {
		t.Fatalf("weak ranking wrong: %+v", weak)
	}
}

func TestInactiveUserIDs_UsesCutoff(t *testing.T) {
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	act := &fakeActivity{inactive: []string{"u1", "u2"}}
	svc := newService(act, fakeProgress{}, fakeQuiz{}, now)

	ids, err := svc.InactiveUserIDs(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 inactive users, got %d", len(ids))
	}
	want := now.AddDate(0, 0, -3)
	if !act.lastSince.Equal(want) {
		t.Fatalf("cutoff wrong: got %v want %v", act.lastSince, want)
	}
}

func TestHandleQuizCompleted_AppendsActivity(t *testing.T) {
	act := &fakeActivity{}
	svc := newService(act, fakeProgress{}, fakeQuiz{}, time.Unix(0, 0))

	evt := quizdomain.NewQuizCompletedEvent("u1", "t1", 90, true, time.Unix(0, 0))
	if err := svc.HandleQuizCompleted(context.Background(), evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(act.appended) != 1 || act.appended[0] != "u1" {
		t.Fatalf("expected activity appended for u1, got %+v", act.appended)
	}
}
