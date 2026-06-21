package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/content/domain"
	"github.com/son-ngo/edu-app/internal/shared/cache"
)

type Service struct {
	repo  domain.Repository
	newID func() string
	cache cache.Cache
	ttl   time.Duration
}

type Option func(*Service)

// WithCache enables cache-aside reads for lessons (read-heavy, rarely changed),
// invalidated when an admin creates a lesson in the topic.
func WithCache(c cache.Cache, ttl time.Duration) Option {
	return func(s *Service) { s.cache = c; s.ttl = ttl }
}

func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, newID: uuid.NewString}
	for _, o := range opts {
		o(s)
	}
	return s
}

func lessonKey(id string) string       { return "content:lesson:" + id }
func lessonsByTopic(tid string) string { return "content:lessons:" + tid }

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
	cache.Invalidate(ctx, s.cache, lessonsByTopic(in.TopicID))
	return lesson, nil
}

func (s *Service) GetLesson(ctx context.Context, id string) (*domain.Lesson, error) {
	return cache.Aside(ctx, s.cache, lessonKey(id), s.ttl, func() (*domain.Lesson, error) {
		return s.repo.GetLesson(ctx, id)
	})
}

func (s *Service) ListByTopic(ctx context.Context, topicID string) ([]domain.Lesson, error) {
	return cache.Aside(ctx, s.cache, lessonsByTopic(topicID), s.ttl, func() ([]domain.Lesson, error) {
		return s.repo.ListByTopic(ctx, topicID)
	})
}
