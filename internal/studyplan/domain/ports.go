package domain

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, p *StudyPlan) error
	GetByID(ctx context.Context, id string) (*StudyPlan, error)
	ListByUser(ctx context.Context, userID string) ([]StudyPlan, error)
}

type TopicSource interface {
	ListTopicIDs(ctx context.Context, subjectID string) ([]string, error)
}

type LevelSource interface {
	Level(ctx context.Context, userID, subjectID string) (string, error)
}

type GoalSource interface {
	PlanWindow(ctx context.Context, userID string) (weeks int, target time.Time, err error)
}

type ReminderEnqueuer interface {
	EnqueueStudyPlanReminder(ctx context.Context, userID, milestoneLabel, idempotencyKey string) error
}
