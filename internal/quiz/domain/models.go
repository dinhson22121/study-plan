// Package domain defines the quiz bounded context: a per-topic practice quiz
// session, its graded result with per-question review, and the ports. Questions
// are read from the question bank via a port to stay decoupled.
package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// MasteryThreshold is the percent score at or above which a quiz is "passed"
// and its topic counts as mastered.
const MasteryThreshold = 80.0

// TestStatus is the lifecycle state of a quiz session.
type TestStatus string

const (
	StatusInProgress TestStatus = "IN_PROGRESS"
	StatusCompleted  TestStatus = "COMPLETED"
)

// QuizSession is the aggregate for one quiz attempt over a topic.
type QuizSession struct {
	ID          string
	UserID      string
	TopicID     string
	Status      TestStatus
	QuestionIDs []string
	CreatedAt   time.Time
}

// NewQuizSession validates and constructs an in-progress session.
func NewQuizSession(id, userID, topicID string, questionIDs []string, now time.Time) (*QuizSession, error) {
	if userID == "" {
		return nil, shared.ErrValidation.WithMessage("user id is required")
	}
	if topicID == "" {
		return nil, shared.ErrValidation.WithMessage("topic id is required")
	}
	if len(questionIDs) == 0 {
		return nil, shared.ErrValidation.WithMessage("no questions available for this topic")
	}
	return &QuizSession{
		ID: id, UserID: userID, TopicID: topicID,
		Status: StatusInProgress, QuestionIDs: questionIDs, CreatedAt: now,
	}, nil
}

// Answer is a student's selected option for a question.
type Answer struct {
	QuestionID string
	OptionID   string
}

// QuestionDetail is the grading/review data for one question, supplied by the
// question-bank source.
type QuestionDetail struct {
	CorrectOptionIDs []string
	Explanation      string
}

// QuestionReview is the post-submit feedback for one question.
type QuestionReview struct {
	QuestionID       string   `json:"question_id"`
	SelectedOptionID string   `json:"selected_option_id"`
	CorrectOptionIDs []string `json:"correct_option_ids"`
	IsCorrect        bool     `json:"is_correct"`
	Explanation      string   `json:"explanation"`
}

// QuizResult is the outcome of a graded session, including per-question review.
type QuizResult struct {
	SessionID    string
	UserID       string
	TopicID      string
	Score        float64
	CorrectCount int
	Total        int
	Passed       bool
	CompletedAt  time.Time
	Reviews      []QuestionReview
}

// Grade scores the session's answers against question details and produces a
// result with per-question review. Only questions in the session count.
func (s *QuizSession) Grade(answers []Answer, details map[string]QuestionDetail, completedAt time.Time) QuizResult {
	selected := make(map[string]string, len(answers))
	for _, a := range answers {
		selected[a.QuestionID] = a.OptionID
	}

	correctSet := func(ids []string) map[string]bool {
		m := make(map[string]bool, len(ids))
		for _, id := range ids {
			m[id] = true
		}
		return m
	}

	var reviews []QuestionReview
	correct := 0
	for _, qid := range s.QuestionIDs {
		detail := details[qid]
		sel := selected[qid]
		isCorrect := sel != "" && correctSet(detail.CorrectOptionIDs)[sel]
		if isCorrect {
			correct++
		}
		reviews = append(reviews, QuestionReview{
			QuestionID:       qid,
			SelectedOptionID: sel,
			CorrectOptionIDs: detail.CorrectOptionIDs,
			IsCorrect:        isCorrect,
			Explanation:      detail.Explanation,
		})
	}

	total := len(s.QuestionIDs)
	score := 0.0
	if total > 0 {
		score = float64(correct) / float64(total) * 100.0
	}
	return QuizResult{
		SessionID: s.ID, UserID: s.UserID, TopicID: s.TopicID,
		Score: score, CorrectCount: correct, Total: total,
		Passed: score >= MasteryThreshold, CompletedAt: completedAt, Reviews: reviews,
	}
}
