package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type Milestone struct {
	ID         string
	WeekNumber int
	TopicIDs   []string
	DueDate    time.Time
}

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

func (p *StudyPlan) FirstMilestone() *Milestone {
	if len(p.Milestones) == 0 {
		return nil
	}
	return &p.Milestones[0]
}
