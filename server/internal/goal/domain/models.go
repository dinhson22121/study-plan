package domain

import (
	"math"
	"strings"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const (
	minScore       = 0.0
	maxScore       = 10.0
	maxDaysPerWeek = 7
)

type SubjectTarget struct {
	SubjectID    string
	CurrentScore float64
	TargetScore  float64
}

type Goal struct {
	UserID           string
	TargetUniversity string
	TargetMajor      string
	TargetDate       time.Time
	HoursPerDay      int
	DaysPerWeek      int
	Subjects         []SubjectTarget
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewGoal(userID, university, major string, targetDate time.Time, hoursPerDay, daysPerWeek int, subjects []SubjectTarget, now time.Time) (*Goal, error) {
	if userID == "" {
		return nil, shared.ErrValidation.WithMessage("user id is required")
	}
	if strings.TrimSpace(university) == "" {
		return nil, shared.ErrValidation.WithMessage("target university is required")
	}
	if !targetDate.After(now) {
		return nil, shared.ErrValidation.WithMessage("target date must be in the future")
	}
	if hoursPerDay <= 0 {
		return nil, shared.ErrValidation.WithMessage("hours per day must be positive")
	}
	if daysPerWeek < 1 || daysPerWeek > maxDaysPerWeek {
		return nil, shared.ErrValidation.WithMessage("days per week must be between 1 and 7")
	}
	if len(subjects) == 0 {
		return nil, shared.ErrValidation.WithMessage("at least one subject target is required")
	}
	for _, s := range subjects {
		if err := validateSubjectTarget(s); err != nil {
			return nil, err
		}
	}
	return &Goal{
		UserID:           userID,
		TargetUniversity: strings.TrimSpace(university),
		TargetMajor:      strings.TrimSpace(major),
		TargetDate:       targetDate,
		HoursPerDay:      hoursPerDay,
		DaysPerWeek:      daysPerWeek,
		Subjects:         subjects,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func validateSubjectTarget(s SubjectTarget) error {
	if s.SubjectID == "" {
		return shared.ErrValidation.WithMessage("subject id is required")
	}
	if s.CurrentScore < minScore || s.CurrentScore > maxScore {
		return shared.ErrValidation.WithMessage("current score must be between 0 and 10")
	}
	if s.TargetScore < minScore || s.TargetScore > maxScore {
		return shared.ErrValidation.WithMessage("target score must be between 0 and 10")
	}
	return nil
}

func (g *Goal) WeeksUntilTarget(now time.Time) int {
	weeks := int(math.Ceil(g.TargetDate.Sub(now).Hours() / (24 * 7)))
	if weeks < 1 {
		return 1
	}
	return weeks
}
