// Package domain defines the studyplan bounded context: a generated per-subject
// learning plan made of weekly milestones, the generation algorithm, and ports.
package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// Milestone is one week's worth of topics with a due date.
type Milestone struct {
	ID         string
	WeekNumber int
	TopicIDs   []string
	DueDate    time.Time
}

// StudyPlan is the aggregate root: an ordered set of weekly milestones covering
// a subject's topics between now and the goal's target date.
type StudyPlan struct {
	ID         string
	UserID     string
	SubjectID  string
	Level      string
	StartDate  time.Time
	TargetDate time.Time
	Milestones []Milestone
	CreatedAt  time.Time
}

// NewStudyPlan validates and constructs a plan.
func NewStudyPlan(id, userID, subjectID, level string, start, target time.Time, milestones []Milestone, now time.Time) (*StudyPlan, error) {
	if userID == "" {
		return nil, shared.ErrValidation.WithMessage("user id is required")
	}
	if subjectID == "" {
		return nil, shared.ErrValidation.WithMessage("subject id is required")
	}
	if len(milestones) == 0 {
		return nil, shared.ErrValidation.WithMessage("plan must have at least one milestone")
	}
	return &StudyPlan{
		ID: id, UserID: userID, SubjectID: subjectID, Level: level,
		StartDate: start, TargetDate: target, Milestones: milestones, CreatedAt: now,
	}, nil
}

// FirstMilestone returns the earliest milestone (week 1), or nil if none.
func (p *StudyPlan) FirstMilestone() *Milestone {
	if len(p.Milestones) == 0 {
		return nil
	}
	return &p.Milestones[0]
}
