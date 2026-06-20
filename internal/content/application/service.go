// Package application contains the content use cases: authoring lessons and
// listing learning materials for a topic.
package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/content/domain"
)

// Service implements the content use cases.
type Service struct {
	repo  domain.Repository
	newID func() string
}

// NewService builds the service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo, newID: uuid.NewString}
}

// ItemInput is a content item in a create request.
type ItemInput struct {
	Kind string
	URL  string
	Body string
}

// CreateLessonInput is the create-lesson command.
type CreateLessonInput struct {
	TopicID    string
	Title      string
	OrderIndex int
	Items      []ItemInput
}

// CreateLesson authors a new lesson with generated ids for it and its items.
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

// GetLesson returns a lesson by id.
func (s *Service) GetLesson(ctx context.Context, id string) (*domain.Lesson, error) {
	return s.repo.GetLesson(ctx, id)
}

// ListByTopic returns the lessons of a topic.
func (s *Service) ListByTopic(ctx context.Context, topicID string) ([]domain.Lesson, error) {
	return s.repo.ListByTopic(ctx, topicID)
}
