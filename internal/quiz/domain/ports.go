package domain

import (
	"context"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type Repository interface {
	SaveSession(ctx context.Context, s *QuizSession) error
	GetSession(ctx context.Context, id string) (*QuizSession, error)

	SaveResultAndComplete(ctx context.Context, r *QuizResult) error

	GetResultForUser(ctx context.Context, sessionID, userID string) (*QuizResult, error)
	ListResultsByUser(ctx context.Context, userID string) ([]QuizResult, error)
}

type QuestionSource interface {
	SampleForTopic(ctx context.Context, topicID string, limit int) ([]string, error)

	Details(ctx context.Context, questionIDs []string) (map[string]QuestionDetail, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, evt shared.DomainEvent) error
}

const EventQuizCompleted = "quiz.completed"

type QuizCompletedEvent struct {
	shared.BaseEvent
	UserID  string
	TopicID string
	Score   float64
	Passed  bool
}

func NewQuizCompletedEvent(userID, topicID string, score float64, passed bool, at time.Time) QuizCompletedEvent {
	return QuizCompletedEvent{
		BaseEvent: shared.BaseEvent{ID: userID, OccurredAtV: at, AggregateV: userID},
		UserID:    userID, TopicID: topicID, Score: score, Passed: passed,
	}
}

func (QuizCompletedEvent) EventName() string { return EventQuizCompleted }
