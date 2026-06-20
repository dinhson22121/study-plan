// Package domain defines the placement bounded context: a per-subject placement
// test, its result and assessed level, and the grading rules. Questions are read
// from the question bank via a port so this context stays decoupled.
package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// Level is the assessed proficiency for a subject.
type Level string

const (
	LevelBeginner     Level = "BEGINNER"
	LevelIntermediate Level = "INTERMEDIATE"
	LevelAdvanced     Level = "ADVANCED"
)

// Score thresholds (percent correct) for level assignment.
const (
	beginnerCeiling     = 40.0 // < 40%  -> BEGINNER
	intermediateCeiling = 75.0 // <= 75% -> INTERMEDIATE, else ADVANCED
)

// LevelFromScore maps a percent-correct score (0–100) to a level.
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

// TestStatus is the lifecycle state of a placement test.
type TestStatus string

const (
	StatusInProgress TestStatus = "IN_PROGRESS"
	StatusCompleted  TestStatus = "COMPLETED"
)

// PlacementTest is the aggregate for one attempt: the snapshot of question ids a
// student is asked to answer for a subject.
type PlacementTest struct {
	ID          string
	UserID      string
	SubjectID   string
	Status      TestStatus
	QuestionIDs []string
	CreatedAt   time.Time
}

// NewPlacementTest validates and constructs an in-progress test.
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

// Answer is a student's selected option for a question.
type Answer struct {
	QuestionID string
	OptionID   string
}

// Grade scores answers against the correct-option sets. Only questions that are
// part of the test count toward the total; an answer is correct when its option
// is in that question's correct set. Returns the percent correct (0–100).
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

// PlacementResult is the outcome of a completed test.
type PlacementResult struct {
	ID          string
	UserID      string
	SubjectID   string
	Score       float64
	Level       Level
	CompletedAt time.Time
}
