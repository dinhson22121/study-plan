package domain

import (
	"context"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// Repository persists placement tests and results.
type Repository interface {
	SaveTest(ctx context.Context, t *PlacementTest) error
	GetTest(ctx context.Context, id string) (*PlacementTest, error)
	// CompleteWithResult stores the result and marks the test COMPLETED in one
	// transaction (atomic submit).
	CompleteWithResult(ctx context.Context, testID string, r *PlacementResult) error
	ListResults(ctx context.Context, userID string) ([]PlacementResult, error)
	// LatestResult returns the most recent result for a user+subject, or
	// ErrNotFound (used by studyplan to read the assessed level).
	LatestResult(ctx context.Context, userID, subjectID string) (*PlacementResult, error)
}

// QuestionSource reads from the question bank to assemble and grade tests,
// keeping placement decoupled from curriculum/question internals.
type QuestionSource interface {
	// SampleForSubject returns up to limit question ids spanning the subject.
	SampleForSubject(ctx context.Context, subjectID string, limit int) ([]string, error)
	// CorrectOptions returns, per question id, the set of correct option ids.
	CorrectOptions(ctx context.Context, questionIDs []string) (map[string]map[string]bool, error)
}

// EventPublisher publishes placement domain events (in-process).
type EventPublisher interface {
	Publish(ctx context.Context, evt shared.DomainEvent) error
}

// Event name for a completed placement.
const EventPlacementCompleted = "placement.completed"

// PlacementCompletedEvent is published when a test is graded.
type PlacementCompletedEvent struct {
	shared.BaseEvent
	UserID    string
	SubjectID string
	Level     Level
	Score     float64
}

// NewPlacementCompletedEvent builds the event.
func NewPlacementCompletedEvent(userID, subjectID string, level Level, score float64, at time.Time) PlacementCompletedEvent {
	return PlacementCompletedEvent{
		BaseEvent: shared.BaseEvent{ID: userID, OccurredAtV: at, AggregateV: userID},
		UserID:    userID, SubjectID: subjectID, Level: level, Score: score,
	}
}

// EventName satisfies shared.DomainEvent.
func (PlacementCompletedEvent) EventName() string { return EventPlacementCompleted }
