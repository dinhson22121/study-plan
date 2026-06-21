package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/curriculum/domain"
	"github.com/son-ngo/edu-app/internal/shared/cache"
)

type Service struct {
	repo  domain.Repository
	newID func() string
	cache cache.Cache
	ttl   time.Duration
}

type Option func(*Service)

// WithCache enables cache-aside reads for the (global, rarely-changing)
// curriculum reference data, invalidated on admin writes.
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

func subjectsKey() string           { return "cur:subjects" }
func chaptersKey(sid string) string { return "cur:chapters:" + sid }
func topicsKey(cid string) string   { return "cur:topics:chapter:" + cid }
func topicKey(id string) string     { return "cur:topic:" + id }

func (s *Service) CreateSubject(ctx context.Context, code, name string, gradeLevel int) (*domain.Subject, error) {
	subject, err := domain.NewSubject(s.newID(), code, name, gradeLevel)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateSubject(ctx, subject); err != nil {
		return nil, err
	}
	cache.Invalidate(ctx, s.cache, subjectsKey())
	return subject, nil
}

func (s *Service) ListSubjects(ctx context.Context) ([]domain.Subject, error) {
	return cache.Aside(ctx, s.cache, subjectsKey(), s.ttl, func() ([]domain.Subject, error) {
		return s.repo.ListSubjects(ctx)
	})
}

func (s *Service) CreateChapter(ctx context.Context, subjectID, title string, orderIndex int) (*domain.Chapter, error) {
	if _, err := s.repo.GetSubject(ctx, subjectID); err != nil {
		return nil, err
	}
	chapter, err := domain.NewChapter(s.newID(), subjectID, title, orderIndex)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateChapter(ctx, chapter); err != nil {
		return nil, err
	}
	cache.Invalidate(ctx, s.cache, chaptersKey(subjectID))
	return chapter, nil
}

func (s *Service) ListChapters(ctx context.Context, subjectID string) ([]domain.Chapter, error) {
	return cache.Aside(ctx, s.cache, chaptersKey(subjectID), s.ttl, func() ([]domain.Chapter, error) {
		return s.repo.ListChaptersBySubject(ctx, subjectID)
	})
}

func (s *Service) CreateTopic(ctx context.Context, chapterID, title string, orderIndex int) (*domain.Topic, error) {
	if _, err := s.repo.GetChapter(ctx, chapterID); err != nil {
		return nil, err
	}
	topic, err := domain.NewTopic(s.newID(), chapterID, title, orderIndex)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateTopic(ctx, topic); err != nil {
		return nil, err
	}
	cache.Invalidate(ctx, s.cache, topicsKey(chapterID))
	return topic, nil
}

func (s *Service) ListTopics(ctx context.Context, chapterID string) ([]domain.Topic, error) {
	return cache.Aside(ctx, s.cache, topicsKey(chapterID), s.ttl, func() ([]domain.Topic, error) {
		return s.repo.ListTopicsByChapter(ctx, chapterID)
	})
}

// ListTopicsBySubject is left uncached: it is consumed cross-module and cannot
// be invalidated cheaply from CreateTopic (which only knows the chapter).
func (s *Service) ListTopicsBySubject(ctx context.Context, subjectID string) ([]domain.Topic, error) {
	return s.repo.ListTopicsBySubject(ctx, subjectID)
}

func (s *Service) GetTopic(ctx context.Context, id string) (*domain.Topic, error) {
	// Topics are immutable once created, so a TTL-only cache never goes stale.
	return cache.Aside(ctx, s.cache, topicKey(id), s.ttl, func() (*domain.Topic, error) {
		return s.repo.GetTopic(ctx, id)
	})
}
