package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/curriculum/domain"
)

type Service struct {
	repo  domain.Repository
	newID func() string
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo, newID: uuid.NewString}
}

func (s *Service) CreateSubject(ctx context.Context, code, name string, gradeLevel int) (*domain.Subject, error) {
	subject, err := domain.NewSubject(s.newID(), code, name, gradeLevel)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateSubject(ctx, subject); err != nil {
		return nil, err
	}
	return subject, nil
}

func (s *Service) ListSubjects(ctx context.Context) ([]domain.Subject, error) {
	return s.repo.ListSubjects(ctx)
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
	return chapter, nil
}

func (s *Service) ListChapters(ctx context.Context, subjectID string) ([]domain.Chapter, error) {
	return s.repo.ListChaptersBySubject(ctx, subjectID)
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
	return topic, nil
}

func (s *Service) ListTopics(ctx context.Context, chapterID string) ([]domain.Topic, error) {
	return s.repo.ListTopicsByChapter(ctx, chapterID)
}

func (s *Service) ListTopicsBySubject(ctx context.Context, subjectID string) ([]domain.Topic, error) {
	return s.repo.ListTopicsBySubject(ctx, subjectID)
}

func (s *Service) GetTopic(ctx context.Context, id string) (*domain.Topic, error) {
	return s.repo.GetTopic(ctx, id)
}
