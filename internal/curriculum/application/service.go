// Package application contains the curriculum use cases: creating and listing
// subjects, chapters, and topics. It validates parent existence before creating
// children so the hierarchy stays consistent.
package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/curriculum/domain"
)

// Service implements the curriculum use cases.
type Service struct {
	repo  domain.Repository
	newID func() string
}

// NewService builds the service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo, newID: uuid.NewString}
}

// CreateSubject creates a new subject.
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

// ListSubjects returns all subjects.
func (s *Service) ListSubjects(ctx context.Context) ([]domain.Subject, error) {
	return s.repo.ListSubjects(ctx)
}

// CreateChapter creates a chapter under an existing subject.
func (s *Service) CreateChapter(ctx context.Context, subjectID, title string, orderIndex int) (*domain.Chapter, error) {
	if _, err := s.repo.GetSubject(ctx, subjectID); err != nil {
		return nil, err // ErrNotFound when subject is missing
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

// ListChapters returns the chapters of a subject.
func (s *Service) ListChapters(ctx context.Context, subjectID string) ([]domain.Chapter, error) {
	return s.repo.ListChaptersBySubject(ctx, subjectID)
}

// CreateTopic creates a topic under an existing chapter.
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

// ListTopics returns the topics of a chapter.
func (s *Service) ListTopics(ctx context.Context, chapterID string) ([]domain.Topic, error) {
	return s.repo.ListTopicsByChapter(ctx, chapterID)
}

// ListTopicsBySubject returns all topics under a subject, ordered for sequential
// study. Used by placement (test assembly) and studyplan (plan generation).
func (s *Service) ListTopicsBySubject(ctx context.Context, subjectID string) ([]domain.Topic, error) {
	return s.repo.ListTopicsBySubject(ctx, subjectID)
}

// GetTopic returns a single topic.
func (s *Service) GetTopic(ctx context.Context, id string) (*domain.Topic, error) {
	return s.repo.GetTopic(ctx, id)
}
