package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/content/domain"
)

type Service struct {
	repo  domain.Repository
	newID func() string
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo, newID: uuid.NewString}
}

type ItemInput struct {
	Kind string
	URL  string
	Body string
}

type CreateLessonInput struct {
	TopicID    string
	Title      string
	OrderIndex int
	Items      []ItemInput
}

func (s *Service) CreateLesson(ctx context.Context, in CreateLessonInput) (*domain.Lesson, error) {
	items := make([]domain.ContentItem, 0, len(in.Items))
	for i, it := range in.Items {
		items = append(items, domain.ContentItem{
			ID:         s.newID(),
			Kind:       domain.ContentKind(it.Kind),
			URL:        it.URL,
			Body:       it.Body,
			OrderIndex: i,
		})
	}
	lesson, err := domain.NewLesson(s.newID(), in.TopicID, in.Title, in.OrderIndex, items)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateLesson(ctx, lesson); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (s *Service) GetLesson(ctx context.Context, id string) (*domain.Lesson, error) {
	return s.repo.GetLesson(ctx, id)
}

func (s *Service) ListByTopic(ctx context.Context, topicID string) ([]domain.Lesson, error) {
	return s.repo.ListByTopic(ctx, topicID)
}
