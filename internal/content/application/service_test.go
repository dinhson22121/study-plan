package application

import (
	"context"
	"errors"
	"testing"

	"github.com/son-ngo/edu-app/internal/content/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeRepo struct {
	lessons map[string]*domain.Lesson
}

func newFakeRepo() *fakeRepo { return &fakeRepo{lessons: map[string]*domain.Lesson{}} }

func (r *fakeRepo) CreateLesson(_ context.Context, l *domain.Lesson) error {
	r.lessons[l.ID] = l
	return nil
}
func (r *fakeRepo) GetLesson(_ context.Context, id string) (*domain.Lesson, error) {
	if l, ok := r.lessons[id]; ok {
		return l, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeRepo) ListByTopic(_ context.Context, topicID string) ([]domain.Lesson, error) {
	var out []domain.Lesson
	for _, l := range r.lessons {
		if l.TopicID == topicID {
			out = append(out, *l)
		}
	}
	return out, nil
}

func TestCreateLesson_GeneratesIDsAndAssignsOrder(t *testing.T) {
	svc := NewService(newFakeRepo())
	lesson, err := svc.CreateLesson(context.Background(), CreateLessonInput{
		TopicID: "t1", Title: "Logarit", Items: []ItemInput{
			{Kind: "PDF", URL: "https://x/p.pdf"},
			{Kind: "NOTE", Body: "ghi chú"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lesson.ID == "" || len(lesson.Items) != 2 {
		t.Fatalf("lesson built wrong: %+v", lesson)
	}
	if lesson.Items[0].ID == "" || lesson.Items[1].OrderIndex != 1 {
		t.Fatalf("item ids/order not assigned: %+v", lesson.Items)
	}
}

func TestCreateLesson_PropagatesValidationError(t *testing.T) {
	svc := NewService(newFakeRepo())
	_, err := svc.CreateLesson(context.Background(), CreateLessonInput{
		TopicID: "t1", Title: "x", Items: []ItemInput{{Kind: "PDF"}}, // PDF missing url
	})
	if !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestListByTopic(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	_, _ = svc.CreateLesson(ctx, CreateLessonInput{TopicID: "t1", Title: "L1"})
	_, _ = svc.CreateLesson(ctx, CreateLessonInput{TopicID: "t1", Title: "L2"})
	_, _ = svc.CreateLesson(ctx, CreateLessonInput{TopicID: "t2", Title: "L3"})

	got, err := svc.ListByTopic(ctx, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lessons for t1, got %d", len(got))
	}
}
