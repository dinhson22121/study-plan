package application

import (
	"context"
	"errors"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/studyplan/domain"
)

type fakeRepo struct{ saved []*domain.StudyPlan }

func (r *fakeRepo) Save(_ context.Context, p *domain.StudyPlan) error {
	r.saved = append(r.saved, p)
	return nil
}
func (r *fakeRepo) GetByID(_ context.Context, id string) (*domain.StudyPlan, error) {
	for _, p := range r.saved {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) ListByUser(_ context.Context, userID string) ([]domain.StudyPlan, error) {
	var out []domain.StudyPlan
	for _, p := range r.saved {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	return out, nil
}

type fakeTopics struct {
	ids []string
	err error
}

func (f *fakeTopics) ListTopicIDs(context.Context, string) ([]string, error) { return f.ids, f.err }

type fakeLevels struct{ level string }

func (f *fakeLevels) Level(context.Context, string, string) (string, error) { return f.level, nil }

type fakeGoals struct {
	weeks  int
	target time.Time
	err    error
}

func (f *fakeGoals) PlanWindow(context.Context, string) (int, time.Time, error) {
	return f.weeks, f.target, f.err
}

type fakeReminder struct{ calls int }

func (r *fakeReminder) EnqueueStudyPlanReminder(context.Context, string, string, string) error {
	r.calls++
	return nil
}

func newSvc(repo *fakeRepo, topics *fakeTopics, levels *fakeLevels, goals *fakeGoals, rem *fakeReminder) *Service {
	s := NewService(repo, topics, levels, goals, rem)
	s.now = func() time.Time { return time.Unix(1_700_000_000, 0).UTC() }
	seq := 0
	s.newID = func() string { seq++; return "id" + string(rune('0'+seq)) }
	return s
}

func TestGeneratePlan_HappyPath(t *testing.T) {
	repo := &fakeRepo{}
	rem := &fakeReminder{}
	target := time.Unix(1_700_000_000, 0).Add(28 * 24 * time.Hour)
	svc := newSvc(repo, &fakeTopics{ids: []string{"t1", "t2", "t3", "t4"}}, &fakeLevels{level: "INTERMEDIATE"}, &fakeGoals{weeks: 4, target: target}, rem)

	plan, err := svc.GeneratePlan(context.Background(), "u1", "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Level != "INTERMEDIATE" || len(plan.Milestones) == 0 {
		t.Fatalf("plan not generated correctly: %+v", plan)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("plan not saved")
	}
	if rem.calls != 1 {
		t.Fatalf("expected first-milestone reminder enqueued, got %d", rem.calls)
	}
}

func TestGeneratePlan_DefaultsLevelWhenNoPlacement(t *testing.T) {
	svc := newSvc(&fakeRepo{}, &fakeTopics{ids: []string{"t1"}}, &fakeLevels{level: ""}, &fakeGoals{weeks: 2, target: time.Unix(2_000_000_000, 0)}, &fakeReminder{})
	plan, err := svc.GeneratePlan(context.Background(), "u1", "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Level != defaultLevel {
		t.Fatalf("expected default level %s, got %s", defaultLevel, plan.Level)
	}
}

func TestGeneratePlan_NoTopicsIsValidationError(t *testing.T) {
	svc := newSvc(&fakeRepo{}, &fakeTopics{ids: nil}, &fakeLevels{}, &fakeGoals{weeks: 2, target: time.Unix(2_000_000_000, 0)}, &fakeReminder{})
	if _, err := svc.GeneratePlan(context.Background(), "u1", "s1"); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for no topics, got %v", err)
	}
}

func TestGeneratePlan_NoGoalPropagatesNotFound(t *testing.T) {
	svc := newSvc(&fakeRepo{}, &fakeTopics{ids: []string{"t1"}}, &fakeLevels{}, &fakeGoals{err: shared.ErrNotFound}, &fakeReminder{})
	if _, err := svc.GeneratePlan(context.Background(), "u1", "s1"); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected not found (set goal first), got %v", err)
	}
}
