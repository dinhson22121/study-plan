package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type Level string

const (
	LevelBeginner     Level = "BEGINNER"
	LevelIntermediate Level = "INTERMEDIATE"
	LevelAdvanced     Level = "ADVANCED"
)

const (
	beginnerCeiling     = 40.0
	intermediateCeiling = 75.0
)

func LevelFromScore(scorePct float64) Level {
	switch {
	case scorePct < beginnerCeiling:
		return LevelBeginner
	case scorePct <= intermediateCeiling:
		return LevelIntermediate
	default:
		return LevelAdvanced
	}
}

type TestStatus string

const (
	StatusInProgress TestStatus = "IN_PROGRESS"
	StatusCompleted  TestStatus = "COMPLETED"
)

type PlacementTest struct {
	ID          string
	UserID      string
	SubjectID   string
	Status      TestStatus
	QuestionIDs []string
	CreatedAt   time.Time
}

func NewPlacementTest(id, userID, subjectID string, questionIDs []string, now time.Time) (*PlacementTest, error) {
	if userID == "" {
		return nil, shared.ErrValidation.WithMessage("user id is required")
	}
	if subjectID == "" {
		return nil, shared.ErrValidation.WithMessage("subject id is required")
	}
	if len(questionIDs) == 0 {
		return nil, shared.ErrValidation.WithMessage("no questions available for this subject")
	}
	return &PlacementTest{
		ID: id, UserID: userID, SubjectID: subjectID,
		Status: StatusInProgress, QuestionIDs: questionIDs, CreatedAt: now,
	}, nil
}

type Answer struct {
	QuestionID string
	OptionID   string
}

func (t *PlacementTest) Grade(answers []Answer, correct map[string]map[string]bool) float64 {
	if len(t.QuestionIDs) == 0 {
		return 0
	}
	inTest := make(map[string]bool, len(t.QuestionIDs))
	for _, qid := range t.QuestionIDs {
		inTest[qid] = true
	}
	got := 0
	for _, a := range answers {
		if !inTest[a.QuestionID] {
			continue
		}
		if correct[a.QuestionID][a.OptionID] {
			got++
		}
	}
	return float64(got) / float64(len(t.QuestionIDs)) * 100.0
}

type PlacementResult struct {
	ID          string
	UserID      string
	SubjectID   string
	Score       float64
	Level       Level
	CompletedAt time.Time
}
