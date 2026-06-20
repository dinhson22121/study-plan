package domain

import (
	"errors"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestLevelFromScore(t *testing.T) {
	cases := []struct {
		score float64
		want  Level
	}{
		{0, LevelBeginner},
		{39.9, LevelBeginner},
		{40, LevelIntermediate},
		{75, LevelIntermediate},
		{75.1, LevelAdvanced},
		{100, LevelAdvanced},
	}
	for _, tc := range cases {
		if got := LevelFromScore(tc.score); got != tc.want {
			t.Fatalf("LevelFromScore(%.1f) = %s, want %s", tc.score, got, tc.want)
		}
	}
}

func TestNewPlacementTest_Validation(t *testing.T) {
	now := time.Unix(0, 0)
	if _, err := NewPlacementTest("id", "", "s", []string{"q"}, now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty user")
	}
	if _, err := NewPlacementTest("id", "u", "s", nil, now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for no questions")
	}
}

func TestPlacementTest_Grade(t *testing.T) {
	test := &PlacementTest{QuestionIDs: []string{"q1", "q2", "q3", "q4"}}
	correct := map[string]map[string]bool{
		"q1": {"a": true},
		"q2": {"b": true},
		"q3": {"c": true},
		"q4": {"d": true},
	}
	answers := []Answer{
		{QuestionID: "q1", OptionID: "a"},
		{QuestionID: "q2", OptionID: "x"},
		{QuestionID: "q3", OptionID: "c"},
		{QuestionID: "q99", OptionID: "z"},
	}

	if score := test.Grade(answers, correct); score != 50.0 {
		t.Fatalf("expected 50.0, got %.2f", score)
	}
}

func TestPlacementTest_GradeEmpty(t *testing.T) {
	test := &PlacementTest{}
	if score := test.Grade(nil, nil); score != 0 {
		t.Fatalf("expected 0 for empty test, got %.2f", score)
	}
}
