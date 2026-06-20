package domain

import (
	"strings"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type QuestionType string

const (
	TypeMCQ      QuestionType = "MCQ"
	TypeFreeText QuestionType = "FREE_TEXT"
)

func (t QuestionType) Valid() bool { return t == TypeMCQ || t == TypeFreeText }

type Difficulty string

const (
	DifficultyEasy   Difficulty = "EASY"
	DifficultyMedium Difficulty = "MEDIUM"
	DifficultyHard   Difficulty = "HARD"
)

func (d Difficulty) Valid() bool {
	return d == DifficultyEasy || d == DifficultyMedium || d == DifficultyHard
}

func ParseDifficulty(s string) (Difficulty, error) {
	d := Difficulty(strings.ToUpper(strings.TrimSpace(s)))
	if !d.Valid() {
		return "", shared.ErrValidation.WithMessage("invalid difficulty: " + s)
	}
	return d, nil
}

type AnswerOption struct {
	ID         string
	Text       string
	IsCorrect  bool
	OrderIndex int
}

type Question struct {
	ID          string
	TopicID     string
	Type        QuestionType
	Stem        string
	Difficulty  Difficulty
	Explanation string
	Options     []AnswerOption
}

func NewQuestion(id, topicID string, qtype QuestionType, stem string, difficulty Difficulty, explanation string, options []AnswerOption) (*Question, error) {
	if topicID == "" {
		return nil, shared.ErrValidation.WithMessage("topic id is required")
	}
	if strings.TrimSpace(stem) == "" {
		return nil, shared.ErrValidation.WithMessage("question stem is required")
	}
	if !qtype.Valid() {
		return nil, shared.ErrValidation.WithMessage("invalid question type")
	}
	if !difficulty.Valid() {
		return nil, shared.ErrValidation.WithMessage("invalid difficulty")
	}

	switch qtype {
	case TypeMCQ:
		if err := validateMCQOptions(options); err != nil {
			return nil, err
		}
	case TypeFreeText:
		if len(options) > 0 {
			return nil, shared.ErrValidation.WithMessage("free-text questions must not have options")
		}
	}

	return &Question{
		ID:          id,
		TopicID:     topicID,
		Type:        qtype,
		Stem:        strings.TrimSpace(stem),
		Difficulty:  difficulty,
		Explanation: strings.TrimSpace(explanation),
		Options:     options,
	}, nil
}

func validateMCQOptions(options []AnswerOption) error {
	if len(options) < 2 {
		return shared.ErrValidation.WithMessage("MCQ requires at least two options")
	}
	correct := 0
	for _, o := range options {
		if strings.TrimSpace(o.Text) == "" {
			return shared.ErrValidation.WithMessage("option text is required")
		}
		if o.IsCorrect {
			correct++
		}
	}
	if correct < 1 {
		return shared.ErrValidation.WithMessage("MCQ requires at least one correct option")
	}
	return nil
}

func (q *Question) IsCorrect(optionID string) bool {
	for _, o := range q.Options {
		if o.ID == optionID {
			return o.IsCorrect
		}
	}
	return false
}
