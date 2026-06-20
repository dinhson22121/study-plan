// Package infrastructure provides the studyplan adapters (bridging curriculum,
// placement, goal, and notification) and the Postgres repository.
package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/son-ngo/edu-app/internal/app"
	curriculumapp "github.com/son-ngo/edu-app/internal/curriculum/application"
	goalapp "github.com/son-ngo/edu-app/internal/goal/application"
	placementapp "github.com/son-ngo/edu-app/internal/placement/application"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// TopicSourceAdapter implements studyplan's TopicSource via curriculum.
type TopicSourceAdapter struct{ curriculum *curriculumapp.Service }

// NewTopicSourceAdapter builds the adapter.
func NewTopicSourceAdapter(c *curriculumapp.Service) *TopicSourceAdapter {
	return &TopicSourceAdapter{curriculum: c}
}

// ListTopicIDs returns the subject's topic ids in curriculum order.
func (a *TopicSourceAdapter) ListTopicIDs(ctx context.Context, subjectID string) ([]string, error) {
	topics, err := a.curriculum.ListTopicsBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(topics))
	for _, t := range topics {
		ids = append(ids, t.ID)
	}
	return ids, nil
}

// LevelSourceAdapter implements studyplan's LevelSource via placement.
type LevelSourceAdapter struct{ placement *placementapp.Service }

// NewLevelSourceAdapter builds the adapter.
func NewLevelSourceAdapter(p *placementapp.Service) *LevelSourceAdapter {
	return &LevelSourceAdapter{placement: p}
}

// Level returns the latest assessed level, or "" when the user hasn't taken a
// placement test for the subject (not an error — generation defaults the level).
func (a *LevelSourceAdapter) Level(ctx context.Context, userID, subjectID string) (string, error) {
	res, err := a.placement.LatestResult(ctx, userID, subjectID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return string(res.Level), nil
}

// GoalSourceAdapter implements studyplan's GoalSource via goal.
type GoalSourceAdapter struct {
	goal *goalapp.Service
	now  func() time.Time
}

// NewGoalSourceAdapter builds the adapter.
func NewGoalSourceAdapter(g *goalapp.Service) *GoalSourceAdapter {
	return &GoalSourceAdapter{goal: g, now: time.Now}
}

// PlanWindow returns the study weeks until the goal's target date and the target.
func (a *GoalSourceAdapter) PlanWindow(ctx context.Context, userID string) (int, time.Time, error) {
	g, err := a.goal.GetGoal(ctx, userID)
	if err != nil {
		return 0, time.Time{}, err
	}
	return g.WeeksUntilTarget(a.now()), g.TargetDate, nil
}

// ReminderAdapter implements studyplan's ReminderEnqueuer via the app.Notifier.
type ReminderAdapter struct{ notifier app.Notifier }

// NewReminderAdapter builds the adapter.
func NewReminderAdapter(n app.Notifier) *ReminderAdapter { return &ReminderAdapter{notifier: n} }

// EnqueueStudyPlanReminder enqueues a STUDY_PLAN notification for a milestone.
func (a *ReminderAdapter) EnqueueStudyPlanReminder(ctx context.Context, userID, milestoneLabel, idempotencyKey string) error {
	if a.notifier == nil {
		return nil // notifications not wired (e.g. tests)
	}
	return a.notifier.EnqueueReminder(ctx, userID, "STUDY_PLAN", "STUDY_PLAN_V1",
		map[string]string{"milestone": milestoneLabel}, idempotencyKey)
}
