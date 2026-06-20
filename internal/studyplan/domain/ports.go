package domain

import (
	"context"
	"time"
)

// Repository persists and retrieves study plans.
type Repository interface {
	Save(ctx context.Context, p *StudyPlan) error
	GetByID(ctx context.Context, id string) (*StudyPlan, error)
	ListByUser(ctx context.Context, userID string) ([]StudyPlan, error)
}

// TopicSource lists a subject's topic ids in curriculum order.
type TopicSource interface {
	ListTopicIDs(ctx context.Context, subjectID string) ([]string, error)
}

// LevelSource returns the assessed level for a user+subject, or "" if none.
type LevelSource interface {
	Level(ctx context.Context, userID, subjectID string) (string, error)
}

// GoalSource provides goal-derived timing for generation.
type GoalSource interface {
	// PlanWindow returns the number of study weeks until the target date and the
	// target date itself. Returns ErrNotFound when the user has no goal.
	PlanWindow(ctx context.Context, userID string) (weeks int, target time.Time, err error)
}

// ReminderEnqueuer triggers a study-plan reminder via the notification pipeline.
type ReminderEnqueuer interface {
	EnqueueStudyPlanReminder(ctx context.Context, userID, milestoneLabel, idempotencyKey string) error
}
