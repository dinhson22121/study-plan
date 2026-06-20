package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/studyplan/domain"
)

const defaultLevel = "BEGINNER"

type Service struct {
	repo     domain.Repository
	topics   domain.TopicSource
	levels   domain.LevelSource
	goals    domain.GoalSource
	reminder domain.ReminderEnqueuer
	now      func() time.Time
	newID    func() string
}

func NewService(repo domain.Repository, topics domain.TopicSource, levels domain.LevelSource, goals domain.GoalSource, reminder domain.ReminderEnqueuer) *Service {
	return &Service{
		repo: repo, topics: topics, levels: levels, goals: goals, reminder: reminder,
		now: time.Now, newID: uuid.NewString,
	}
}

func (s *Service) GeneratePlan(ctx context.Context, userID, subjectID string) (*domain.StudyPlan, error) {
	weeks, target, err := s.goals.PlanWindow(ctx, userID)
	if err != nil {
		return nil, err
	}

	topicIDs, err := s.topics.ListTopicIDs(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	if len(topicIDs) == 0 {
		return nil, shared.ErrValidation.WithMessage("subject has no topics to plan")
	}

	level, err := s.levels.Level(ctx, userID, subjectID)
	if err != nil {
		return nil, err
	}
	if level == "" {
		level = defaultLevel
	}

	now := s.now()
	milestones := domain.GenerateMilestones(topicIDs, weeks, now, s.newID)
	plan, err := domain.NewStudyPlan(s.newID(), userID, subjectID, level, now, target, milestones, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, plan); err != nil {
		return nil, err
	}

	if first := plan.FirstMilestone(); first != nil {
		label := fmt.Sprintf("Tuần %d", first.WeekNumber)
		idemKey := plan.ID + "-w" + fmt.Sprint(first.WeekNumber)
		_ = s.reminder.EnqueueStudyPlanReminder(ctx, userID, label, idemKey)
	}

	return plan, nil
}

func (s *Service) GetPlan(ctx context.Context, id string) (*domain.StudyPlan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListPlans(ctx context.Context, userID string) ([]domain.StudyPlan, error) {
	return s.repo.ListByUser(ctx, userID)
}
