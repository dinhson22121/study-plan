package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/placement/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const defaultNumQuestions = 20

type Service struct {
	repo   domain.Repository
	source domain.QuestionSource
	bus    domain.EventPublisher
	now    func() time.Time
	newID  func() string
}

func NewService(repo domain.Repository, source domain.QuestionSource, bus domain.EventPublisher) *Service {
	return &Service{repo: repo, source: source, bus: bus, now: time.Now, newID: uuid.NewString}
}

func (s *Service) StartTest(ctx context.Context, userID, subjectID string, numQuestions int) (*domain.PlacementTest, error) {
	if numQuestions <= 0 {
		numQuestions = defaultNumQuestions
	}
	questionIDs, err := s.source.SampleForSubject(ctx, subjectID, numQuestions)
	if err != nil {
		return nil, err
	}
	test, err := domain.NewPlacementTest(s.newID(), userID, subjectID, questionIDs, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveTest(ctx, test); err != nil {
		return nil, err
	}
	return test, nil
}

type AnswerInput struct {
	QuestionID string
	OptionID   string
}

func (s *Service) SubmitTest(ctx context.Context, testID, userID string, answers []AnswerInput) (*domain.PlacementResult, error) {
	test, err := s.repo.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}
	if test.UserID != userID {
		return nil, shared.ErrForbidden.WithMessage("test belongs to another user")
	}
	if test.Status != domain.StatusInProgress {
		return nil, shared.ErrConflict.WithMessage("test already submitted")
	}

	correct, err := s.source.CorrectOptions(ctx, test.QuestionIDs)
	if err != nil {
		return nil, err
	}

	domainAnswers := make([]domain.Answer, 0, len(answers))
	for _, a := range answers {
		domainAnswers = append(domainAnswers, domain.Answer{QuestionID: a.QuestionID, OptionID: a.OptionID})
	}
	score := test.Grade(domainAnswers, correct)
	level := domain.LevelFromScore(score)

	result := &domain.PlacementResult{
		ID: s.newID(), UserID: userID, SubjectID: test.SubjectID,
		Score: score, Level: level, CompletedAt: s.now(),
	}
	if err := s.repo.CompleteWithResult(ctx, testID, result); err != nil {
		return nil, err
	}

	evt := domain.NewPlacementCompletedEvent(userID, test.SubjectID, level, score, s.now())
	if err := s.bus.Publish(ctx, evt); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	return result, nil
}

func (s *Service) ListResults(ctx context.Context, userID string) ([]domain.PlacementResult, error) {
	return s.repo.ListResults(ctx, userID)
}

func (s *Service) LatestResult(ctx context.Context, userID, subjectID string) (*domain.PlacementResult, error) {
	return s.repo.LatestResult(ctx, userID, subjectID)
}
