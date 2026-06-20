package domain

import (
	"context"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type Repository interface {
	SaveTest(ctx context.Context, t *PlacementTest) error
	GetTest(ctx context.Context, id string) (*PlacementTest, error)

	CompleteWithResult(ctx context.Context, testID string, r *PlacementResult) error
	ListResults(ctx context.Context, userID string) ([]PlacementResult, error)

	LatestResult(ctx context.Context, userID, subjectID string) (*PlacementResult, error)
}

type QuestionSource interface {
	SampleForSubject(ctx context.Context, subjectID string, limit int) ([]string, error)

	CorrectOptions(ctx context.Context, questionIDs []string) (map[string]map[string]bool, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, evt shared.DomainEvent) error
}

const EventPlacementCompleted = "placement.completed"

type PlacementCompletedEvent struct {
	shared.BaseEvent
	UserID    string
	SubjectID string
	Level     Level
	Score     float64
}

func NewPlacementCompletedEvent(userID, subjectID string, level Level, score float64, at time.Time) PlacementCompletedEvent {
	return PlacementCompletedEvent{
		BaseEvent: shared.BaseEvent{ID: userID, OccurredAtV: at, AggregateV: userID},
		UserID:    userID, SubjectID: subjectID, Level: level, Score: score,
	}
}

func (PlacementCompletedEvent) EventName() string { return EventPlacementCompleted }
