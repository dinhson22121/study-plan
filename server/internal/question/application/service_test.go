package application

import (
	"context"
	"errors"
	"testing"

	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	created []*domain.Question
	byID    map[string]*domain.Question
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byID: map[string]*domain.Question{}} }

func (r *fakeRepo) Create(_ context.Context, q *domain.Question) error {
	r.created = append(r.created, q)
	r.byID[q.ID] = q
	return nil
}
func (r *fakeRepo) GetByID(_ context.Context, id string) (*domain.Question, error) {
	if q, ok := r.byID[id]; ok {
		return q, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) List(_ context.Context, f domain.ListFilter) ([]domain.Question, error) {
	var out []domain.Question
	for _, q := range r.byID {
		if q.TopicID == f.TopicID && (f.Difficulty == "" || q.Difficulty == f.Difficulty) {
			out = append(out, *q)
		}
	}
	return out, nil
}

func validMCQ() CreateInput {
	return CreateInput{
		TopicID: "t1", Type: "MCQ", Stem: "2+2=?", Difficulty: "easy", Explanation: "math",
		Options: []OptionInput{{Text: "3"}, {Text: "4", IsCorrect: true}},
	}
}

func TestCreate_GeneratesIDsAndPersists(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	q, err := svc.Create(context.Background(), validMCQ())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.ID == "" || len(q.Options) != 2 {
		t.Fatalf("question not built correctly: %+v", q)
	}
	for _, o := range q.Options {
		if o.ID == "" {
			t.Fatalf("option id not generated")
		}
	}
	if q.Difficulty != domain.DifficultyEasy {
		t.Fatalf("difficulty not parsed: %s", q.Difficulty)
	}
}

func TestCreate_RejectsInvalidDifficulty(t *testing.T) {
	svc := NewService(newFakeRepo())
	in := validMCQ()
	in.Difficulty = "impossible"
	if _, err := svc.Create(context.Background(), in); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreate_RejectsInvalidMCQ(t *testing.T) {
	svc := NewService(newFakeRepo())
	in := validMCQ()
	in.Options = []OptionInput{{Text: "only one"}}
	if _, err := svc.Create(context.Background(), in); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for single option, got %v", err)
	}
}

func TestList_FiltersByDifficulty(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	ctx := context.Background()
	easy := validMCQ()
	hard := validMCQ()
	hard.Difficulty = "hard"
	_, _ = svc.Create(ctx, easy)
	_, _ = svc.Create(ctx, hard)

	got, err := svc.List(ctx, "t1", "hard", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Difficulty != domain.DifficultyHard {
		t.Fatalf("expected 1 hard question, got %d", len(got))
	}

	if _, err := svc.List(ctx, "t1", "bogus", 10); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error for bad difficulty filter")
	}
}
