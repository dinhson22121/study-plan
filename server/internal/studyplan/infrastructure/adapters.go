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

type TopicSourceAdapter struct{ curriculum *curriculumapp.Service }

func NewTopicSourceAdapter(c *curriculumapp.Service) *TopicSourceAdapter {
	return &TopicSourceAdapter{curriculum: c}
}

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

type LevelSourceAdapter struct{ placement *placementapp.Service }

func NewLevelSourceAdapter(p *placementapp.Service) *LevelSourceAdapter {
	return &LevelSourceAdapter{placement: p}
}

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

type GoalSourceAdapter struct {
	goal *goalapp.Service
	now  func() time.Time
}

func NewGoalSourceAdapter(g *goalapp.Service) *GoalSourceAdapter {
	return &GoalSourceAdapter{goal: g, now: time.Now}
}

func (a *GoalSourceAdapter) PlanWindow(ctx context.Context, userID string) (int, time.Time, error) {
	g, err := a.goal.GetGoal(ctx, userID)
	if err != nil {
		return 0, time.Time{}, err
	}
	return g.WeeksUntilTarget(a.now()), g.TargetDate, nil
}

type ReminderAdapter struct{ notifier app.Notifier }

func NewReminderAdapter(n app.Notifier) *ReminderAdapter { return &ReminderAdapter{notifier: n} }

func (a *ReminderAdapter) EnqueueStudyPlanReminder(ctx context.Context, userID, milestoneLabel, idempotencyKey string) error {
	if a.notifier == nil {
		return nil
	}
	return a.notifier.EnqueueReminder(ctx, userID, "STUDY_PLAN", "STUDY_PLAN_V1",
		map[string]string{"milestone": milestoneLabel}, idempotencyKey)
}
