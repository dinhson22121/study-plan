// Package application contains the goal use cases: setting and reading a user's
// learning goal.
package application

import (
	"context"
	"time"

	"github.com/son-ngo/edu-app/internal/goal/domain"
)

// Service implements the goal use cases.
type Service struct {
	repo domain.Repository
	now  func() time.Time
}

// Option customizes a Service.
type Option func(*Service)

// WithClock overrides the time source (tests).
func WithClock(now func() time.Time) Option { return func(s *Service) { s.now = now } }

// NewService builds the service.
func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, o := range opts {
		o(s)
	}
	return s
}

// SubjectTargetInput is a per-subject score target in a set-goal command.
type SubjectTargetInput struct {
	SubjectID    string
	CurrentScore float64
	TargetScore  float64
}

// SetGoalInput is the set-goal command (create or replace).
type SetGoalInput struct {
	UserID           string
	TargetUniversity string
	TargetMajor      string
	TargetDate       time.Time
	HoursPerDay      int
	DaysPerWeek      int
	Subjects         []SubjectTargetInput
}

// SetGoal creates or replaces the user's goal.
func (s *Service) SetGoal(ctx context.Context, in SetGoalInput) (*domain.Goal, error) {
	subjects := make([]domain.SubjectTarget, 0, len(in.Subjects))
	for _, st := range in.Subjects {
		subjects = append(subjects, domain.SubjectTarget{
			SubjectID: st.SubjectID, CurrentScore: st.CurrentScore, TargetScore: st.TargetScore,
		})
	}
	goal, err := domain.NewGoal(in.UserID, in.TargetUniversity, in.TargetMajor, in.TargetDate,
		in.HoursPerDay, in.DaysPerWeek, subjects, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Upsert(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

// GetGoal returns the user's goal.
func (s *Service) GetGoal(ctx context.Context, userID string) (*domain.Goal, error) {
	return s.repo.GetByUserID(ctx, userID)
}
