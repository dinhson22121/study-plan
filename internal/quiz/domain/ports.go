package domain

import (
	"context"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// Repository persists quiz sessions and results.
type Repository interface {
	SaveSession(ctx context.Context, s *QuizSession) error
	GetSession(ctx context.Context, id string) (*QuizSession, error)
	// SaveResultAndComplete stores the result with its review and marks the
	// session COMPLETED in a single transaction (atomic submit).
	SaveResultAndComplete(ctx context.Context, r *QuizResult) error
	// GetResultForUser returns a result only if it belongs to userID (the
	// ownership check is enforced in the query, not the handler).
	GetResultForUser(ctx context.Context, sessionID, userID string) (*QuizResult, error)
	ListResultsByUser(ctx context.Context, userID string) ([]QuizResult, error)
}

// QuestionSource reads from the question bank to assemble and grade quizzes.
type QuestionSource interface {
	// SampleForTopic returns up to limit question ids for a topic.
	SampleForTopic(ctx context.Context, topicID string, limit int) ([]string, error)
	// Details returns grading/review data (correct options + explanation) per id.
	Details(ctx context.Context, questionIDs []string) (map[string]QuestionDetail, error)
}

// EventPublisher publishes quiz domain events (in-process).
type EventPublisher interface {
	Publish(ctx context.Context, evt shared.DomainEvent) error
}

// EventQuizCompleted is the name of the event progress/analytics subscribe to.
const EventQuizCompleted = "quiz.completed"

// QuizCompletedEvent is published when a quiz is graded.
type QuizCompletedEvent struct {
	shared.BaseEvent
	UserID  string
	TopicID string
	Score   float64
	Passed  bool
}

// NewQuizCompletedEvent builds the event.
func NewQuizCompletedEvent(userID, topicID string, score float64, passed bool, at time.Time) QuizCompletedEvent {
	return QuizCompletedEvent{
		BaseEvent: shared.BaseEvent{ID: userID, OccurredAtV: at, AggregateV: userID},
		UserID:    userID, TopicID: topicID, Score: score, Passed: passed,
	}
}

// EventName satisfies shared.DomainEvent.
func (QuizCompletedEvent) EventName() string { return EventQuizCompleted }
