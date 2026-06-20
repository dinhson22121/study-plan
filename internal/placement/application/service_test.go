package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/son-ngo/edu-app/internal/placement/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	tests   map[string]*domain.PlacementTest
	results []*domain.PlacementResult
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{tests: map[string]*domain.PlacementTest{}}
}
func (r *fakeRepo) SaveTest(_ context.Context, t *domain.PlacementTest) error {
	cp := *t
	r.tests[t.ID] = &cp
	return nil
}
func (r *fakeRepo) GetTest(_ context.Context, id string) (*domain.PlacementTest, error) {
	if t, ok := r.tests[id]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) CompleteWithResult(_ context.Context, testID string, res *domain.PlacementResult) error {
	t, ok := r.tests[testID]
	if !ok {
		return shared.ErrNotFound
	}
	t.Status = domain.StatusCompleted
	r.results = append(r.results, res)
	return nil
}
func (r *fakeRepo) ListResults(_ context.Context, userID string) ([]domain.PlacementResult, error) {
	var out []domain.PlacementResult
	for _, res := range r.results {
		if res.UserID == userID {
			out = append(out, *res)
		}
	}
	return out, nil
}
func (r *fakeRepo) LatestResult(_ context.Context, userID, subjectID string) (*domain.PlacementResult, error) {
	for i := len(r.results) - 1; i >= 0; i-- {
		if r.results[i].UserID == userID && r.results[i].SubjectID == subjectID {
			return r.results[i], nil
		}
	}
	return nil, shared.ErrNotFound
}

type fakeSource struct {
	ids     []string
	correct map[string]map[string]bool
}

func (s *fakeSource) SampleForSubject(_ context.Context, _ string, limit int) ([]string, error) {
	if limit < len(s.ids) {
		return s.ids[:limit], nil
	}
	return s.ids, nil
}
func (s *fakeSource) CorrectOptions(_ context.Context, _ []string) (map[string]map[string]bool, error) {
	return s.correct, nil
}

type fakeBus struct{ events []shared.DomainEvent }

func (b *fakeBus) Publish(_ context.Context, e shared.DomainEvent) error {
	b.events = append(b.events, e)
	return nil
}

func newService(repo *fakeRepo, src *fakeSource, bus *fakeBus) *Service {
	s := NewService(repo, src, bus)
	s.now = func() time.Time { return time.Unix(1000, 0).UTC() }
	seq := 0
	s.newID = func() string { seq++; return "id" + string(rune('0'+seq)) }
	return s
}

func fourQuestionSource() *fakeSource {
	return &fakeSource{
		ids: []string{"q1", "q2", "q3", "q4"},
		correct: map[string]map[string]bool{
			"q1": {"a": true}, "q2": {"b": true}, "q3": {"c": true}, "q4": {"d": true},
		},
	}
}

func TestStartTest_AssemblesAndSaves(t *testing.T) {
	repo, src, bus := newFakeRepo(), fourQuestionSource(), &fakeBus{}
	svc := newService(repo, src, bus)

	test, err := svc.StartTest(context.Background(), "u1", "s1", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(test.QuestionIDs) != 4 || test.Status != domain.StatusInProgress {
		t.Fatalf("test not assembled: %+v", test)
	}
	if _, ok := repo.tests[test.ID]; !ok {
		t.Fatalf("test not saved")
	}
}

func TestSubmitTest_GradesAndPublishesEvent(t *testing.T) {
	repo, src, bus := newFakeRepo(), fourQuestionSource(), &fakeBus{}
	svc := newService(repo, src, bus)
	test, _ := svc.StartTest(context.Background(), "u1", "s1", 4)

	result, err := svc.SubmitTest(context.Background(), test.ID, "u1", []AnswerInput{
		{QuestionID: "q1", OptionID: "a"},
		{QuestionID: "q2", OptionID: "b"},
		{QuestionID: "q3", OptionID: "c"},
		{QuestionID: "q4", OptionID: "WRONG"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score != 75.0 || result.Level != domain.LevelIntermediate {
		t.Fatalf("grading wrong: score=%.1f level=%s", result.Score, result.Level)
	}
	if repo.tests[test.ID].Status != domain.StatusCompleted {
		t.Fatalf("test not marked completed")
	}
	if len(bus.events) != 1 || bus.events[0].EventName() != domain.EventPlacementCompleted {
		t.Fatalf("expected PlacementCompletedEvent, got %+v", bus.events)
	}
}

func TestSubmitTest_RejectsOtherUser(t *testing.T) {
	repo, src, bus := newFakeRepo(), fourQuestionSource(), &fakeBus{}
	svc := newService(repo, src, bus)
	test, _ := svc.StartTest(context.Background(), "u1", "s1", 4)

	_, err := svc.SubmitTest(context.Background(), test.ID, "intruder", nil)
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestSubmitTest_RejectsDoubleSubmit(t *testing.T) {
	repo, src, bus := newFakeRepo(), fourQuestionSource(), &fakeBus{}
	svc := newService(repo, src, bus)
	test, _ := svc.StartTest(context.Background(), "u1", "s1", 4)
	_, _ = svc.SubmitTest(context.Background(), test.ID, "u1", []AnswerInput{{QuestionID: "q1", OptionID: "a"}})

	_, err := svc.SubmitTest(context.Background(), test.ID, "u1", nil)
	if !errors.Is(err, shared.ErrConflict) {
		t.Fatalf("expected conflict on double submit, got %v", err)
	}
}
