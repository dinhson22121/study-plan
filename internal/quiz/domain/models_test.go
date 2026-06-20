package domain

import (
	"errors"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestNewQuizSession_Validation(t *testing.T) {
	now := time.Unix(0, 0)
	if _, err := NewQuizSession("id", "", "t", []string{"q"}, now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty user")
	}
	if _, err := NewQuizSession("id", "u", "t", nil, now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for no questions")
	}
}

func TestQuizSession_GradeScoresAndReviews(t *testing.T) {
	s := &QuizSession{ID: "s1", UserID: "u1", TopicID: "t1", QuestionIDs: []string{"q1", "q2", "q3", "q4", "q5"}}
	details := map[string]QuestionDetail{
		"q1": {CorrectOptionIDs: []string{"a"}, Explanation: "e1"},
		"q2": {CorrectOptionIDs: []string{"b"}, Explanation: "e2"},
		"q3": {CorrectOptionIDs: []string{"c"}, Explanation: "e3"},
		"q4": {CorrectOptionIDs: []string{"d"}, Explanation: "e4"},
		"q5": {CorrectOptionIDs: []string{"e"}, Explanation: "e5"},
	}
	answers := []Answer{
		{QuestionID: "q1", OptionID: "a"}, // correct
		{QuestionID: "q2", OptionID: "b"}, // correct
		{QuestionID: "q3", OptionID: "c"}, // correct
		{QuestionID: "q4", OptionID: "c"}, // wrong
		// q5 unanswered
	}
	res := s.Grade(answers, details, time.Unix(100, 0))

	if res.Total != 5 || res.CorrectCount != 3 {
		t.Fatalf("expected 3/5, got %d/%d", res.CorrectCount, res.Total)
	}
	if res.Score != 60.0 {
		t.Fatalf("expected 60.0, got %.2f", res.Score)
	}
	if res.Passed {
		t.Fatalf("60%% should not pass (threshold 80%%)")
	}
	if len(res.Reviews) != 5 {
		t.Fatalf("expected 5 reviews, got %d", len(res.Reviews))
	}
	// q5 unanswered -> not correct, empty selected, explanation present
	last := res.Reviews[4]
	if last.QuestionID != "q5" || last.IsCorrect || last.SelectedOptionID != "" || last.Explanation != "e5" {
		t.Fatalf("unanswered review wrong: %+v", last)
	}
}

func TestQuizSession_GradePass(t *testing.T) {
	s := &QuizSession{QuestionIDs: []string{"q1", "q2", "q3", "q4", "q5"}}
	details := map[string]QuestionDetail{
		"q1": {CorrectOptionIDs: []string{"a"}}, "q2": {CorrectOptionIDs: []string{"a"}},
		"q3": {CorrectOptionIDs: []string{"a"}}, "q4": {CorrectOptionIDs: []string{"a"}},
		"q5": {CorrectOptionIDs: []string{"a"}},
	}
	var answers []Answer
	for _, q := range s.QuestionIDs {
		answers = append(answers, Answer{QuestionID: q, OptionID: "a"})
	}
	res := s.Grade(answers, details, time.Unix(0, 0))
	if res.Score != 100.0 || !res.Passed {
		t.Fatalf("expected 100%% pass, got %.2f passed=%v", res.Score, res.Passed)
	}
}
