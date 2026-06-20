package domain

import (
	"errors"
	"testing"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func mcqOptions(correct int) []AnswerOption {
	return []AnswerOption{
		{ID: "o1", Text: "A", IsCorrect: correct == 0},
		{ID: "o2", Text: "B", IsCorrect: correct == 1},
	}
}

func TestNewQuestion_MCQHappyPath(t *testing.T) {
	q, err := NewQuestion("id", "topic1", TypeMCQ, "1+1=?", DifficultyEasy, "basic", mcqOptions(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !q.IsCorrect("o2") || q.IsCorrect("o1") {
		t.Fatalf("IsCorrect wrong")
	}
}

func TestNewQuestion_MCQValidation(t *testing.T) {
	// fewer than two options
	if _, err := NewQuestion("id", "t", TypeMCQ, "s", DifficultyEasy, "", []AnswerOption{{ID: "o", Text: "A", IsCorrect: true}}); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for <2 options")
	}
	// no correct option
	if _, err := NewQuestion("id", "t", TypeMCQ, "s", DifficultyEasy, "", mcqOptions(-1)); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for no correct option")
	}
	// empty option text
	bad := []AnswerOption{{ID: "o1", Text: "", IsCorrect: true}, {ID: "o2", Text: "B"}}
	if _, err := NewQuestion("id", "t", TypeMCQ, "s", DifficultyEasy, "", bad); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty option text")
	}
}

func TestNewQuestion_FreeTextRejectsOptions(t *testing.T) {
	if _, err := NewQuestion("id", "t", TypeFreeText, "explain X", DifficultyHard, "", mcqOptions(1)); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error: free-text must not have options")
	}
	if _, err := NewQuestion("id", "t", TypeFreeText, "explain X", DifficultyHard, "", nil); err != nil {
		t.Fatalf("free-text with no options should be valid, got %v", err)
	}
}

func TestNewQuestion_RejectsBadTypeAndDifficulty(t *testing.T) {
	if _, err := NewQuestion("id", "t", QuestionType("ESSAY"), "s", DifficultyEasy, "", nil); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for bad type")
	}
	if _, err := NewQuestion("id", "t", TypeFreeText, "s", Difficulty("TRIVIAL"), "", nil); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for bad difficulty")
	}
}

func TestParseDifficulty(t *testing.T) {
	d, err := ParseDifficulty(" hard ")
	if err != nil || d != DifficultyHard {
		t.Fatalf("expected HARD, got %v / %v", d, err)
	}
	if _, err := ParseDifficulty("x"); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error")
	}
}
