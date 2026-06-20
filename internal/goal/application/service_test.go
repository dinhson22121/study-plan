package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/son-ngo/edu-app/internal/goal/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	goals map[string]*domain.Goal
}

func newFakeRepo() *fakeRepo { return &fakeRepo{goals: map[string]*domain.Goal{}} }

func (r *fakeRepo) Upsert(_ context.Context, g *domain.Goal) error {
	cp := *g
	r.goals[g.UserID] = &cp
	return nil
}
func (r *fakeRepo) GetByUserID(_ context.Context, userID string) (*domain.Goal, error) {
	if g, ok := r.goals[userID]; ok {
		return g, nil
	}
	return nil, shared.ErrNotFound
}

func fixedNow() time.Time { return time.Unix(1_700_000_000, 0).UTC() }

func newService(repo *fakeRepo) *Service {
	return NewService(repo, WithClock(fixedNow))
}

func validInput() SetGoalInput {
	return SetGoalInput{
		UserID: "u1", TargetUniversity: "HUST", TargetMajor: "CNTT",
		TargetDate:  fixedNow().Add(60 * 24 * time.Hour),
		HoursPerDay: 2, DaysPerWeek: 5,
		Subjects: []SubjectTargetInput{{SubjectID: "s1", CurrentScore: 5, TargetScore: 8}},
	}
}

func TestSetGoal_CreatesAndReads(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo)

	g, err := svc.SetGoal(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Subjects) != 1 {
		t.Fatalf("expected 1 subject target")
	}
	got, err := svc.GetGoal(context.Background(), "u1")
	if err != nil || got.TargetUniversity != "HUST" {
		t.Fatalf("get goal mismatch: %+v / %v", got, err)
	}
}

func TestSetGoal_IsUpsert(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo)
	_, _ = svc.SetGoal(context.Background(), validInput())

	in := validInput()
	in.TargetUniversity = "VNU"
	if _, err := svc.SetGoal(context.Background(), in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetGoal(context.Background(), "u1")
	if got.TargetUniversity != "VNU" {
		t.Fatalf("expected goal to be replaced, got %s", got.TargetUniversity)
	}
}

func TestSetGoal_ValidationPropagates(t *testing.T) {
	svc := newService(newFakeRepo())
	in := validInput()
	in.TargetDate = fixedNow().Add(-time.Hour) // past
	if _, err := svc.SetGoal(context.Background(), in); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestGetGoal_NotFound(t *testing.T) {
	svc := newService(newFakeRepo())
	if _, err := svc.GetGoal(context.Background(), "ghost"); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
