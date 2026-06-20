package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/quiz/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const defaultNumQuestions = 10

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

func (s *Service) StartQuiz(ctx context.Context, userID, topicID string, numQuestions int) (*domain.QuizSession, error) {
	if numQuestions <= 0 {
		numQuestions = defaultNumQuestions
	}
	ids, err := s.source.SampleForTopic(ctx, topicID, numQuestions)
	if err != nil {
		return nil, err
	}
	session, err := domain.NewQuizSession(s.newID(), userID, topicID, ids, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

type AnswerInput struct {
	QuestionID string
	OptionID   string
}

func (s *Service) SubmitQuiz(ctx context.Context, sessionID, userID string, answers []AnswerInput) (*domain.QuizResult, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, shared.ErrForbidden.WithMessage("quiz belongs to another user")
	}
	if session.Status != domain.StatusInProgress {
		return nil, shared.ErrConflict.WithMessage("quiz already submitted")
	}

	details, err := s.source.Details(ctx, session.QuestionIDs)
	if err != nil {
		return nil, err
	}

	domainAnswers := make([]domain.Answer, 0, len(answers))
	for _, a := range answers {
		domainAnswers = append(domainAnswers, domain.Answer{QuestionID: a.QuestionID, OptionID: a.OptionID})
	}
	result := session.Grade(domainAnswers, details, s.now())

	if err := s.repo.SaveResultAndComplete(ctx, &result); err != nil {
		return nil, err
	}

	evt := domain.NewQuizCompletedEvent(userID, session.TopicID, result.Score, result.Passed, s.now())
	if err := s.bus.Publish(ctx, evt); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &result, nil
}

func (s *Service) GetResult(ctx context.Context, sessionID, userID string) (*domain.QuizResult, error) {
	return s.repo.GetResultForUser(ctx, sessionID, userID)
}

func (s *Service) ListResults(ctx context.Context, userID string) ([]domain.QuizResult, error) {
	return s.repo.ListResultsByUser(ctx, userID)
}
