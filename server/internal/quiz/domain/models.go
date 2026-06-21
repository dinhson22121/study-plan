package domain

import (
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const MasteryThreshold = 80.0

type TestStatus string

const (
	StatusInProgress TestStatus = "IN_PROGRESS"
	StatusCompleted  TestStatus = "COMPLETED"
)

type QuizSession struct {
	ID          string
	UserID      string
	TopicID     string
	Status      TestStatus
	QuestionIDs []string
	CreatedAt   time.Time
}

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

type Answer struct {
	QuestionID string
	OptionID   string
}

type QuestionDetail struct {
	CorrectOptionIDs []string
	Explanation      string
}

type QuestionReview struct {
	QuestionID       string   `json:"question_id"`
	SelectedOptionID string   `json:"selected_option_id"`
	CorrectOptionIDs []string `json:"correct_option_ids"`
	IsCorrect        bool     `json:"is_correct"`
	Explanation      string   `json:"explanation"`
}

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
