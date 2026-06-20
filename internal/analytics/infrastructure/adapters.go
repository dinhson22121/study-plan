// Package infrastructure provides the analytics adapters (progress + quiz
// readers) and the Postgres activity repository.
package infrastructure

import (
	"context"

	"github.com/son-ngo/edu-app/internal/analytics/domain"
	progressapp "github.com/son-ngo/edu-app/internal/progress/application"
	progressdomain "github.com/son-ngo/edu-app/internal/progress/domain"
	quizapp "github.com/son-ngo/edu-app/internal/quiz/application"
)

// ProgressReaderAdapter implements analytics's ProgressReader over progress.
type ProgressReaderAdapter struct{ progress *progressapp.Service }

// NewProgressReaderAdapter builds the adapter.
func NewProgressReaderAdapter(p *progressapp.Service) *ProgressReaderAdapter {
	return &ProgressReaderAdapter{progress: p}
}

// Snapshot maps a progress overview into the analytics snapshot shape.
func (a *ProgressReaderAdapter) Snapshot(ctx context.Context, userID string) (domain.ProgressSnapshot, error) {
	ov, err := a.progress.GetOverview(ctx, userID)
	if err != nil {
		return domain.ProgressSnapshot{}, err
	}
	topics := make([]domain.TopicStat, 0, len(ov.Topics))
	for _, t := range ov.Topics {
		topics = append(topics, domain.TopicStat{
			TopicID:   t.TopicID,
			Completed: t.Status == progressdomain.StatusCompleted,
			BestScore: t.BestScore,
		})
	}
	return domain.ProgressSnapshot{
		CurrentStreak:   ov.CurrentStreak,
		LongestStreak:   ov.LongestStreak,
		TopicsTotal:     ov.TopicsTotal,
		TopicsCompleted: ov.TopicsDone,
		Topics:          topics,
	}, nil
}

// QuizReaderAdapter implements analytics's QuizReader over quiz.
type QuizReaderAdapter struct{ quiz *quizapp.Service }

// NewQuizReaderAdapter builds the adapter.
func NewQuizReaderAdapter(q *quizapp.Service) *QuizReaderAdapter { return &QuizReaderAdapter{quiz: q} }

// Scores returns the user's quiz result scores.
func (a *QuizReaderAdapter) Scores(ctx context.Context, userID string) ([]float64, error) {
	results, err := a.quiz.ListResults(ctx, userID)
	if err != nil {
		return nil, err
	}
	scores := make([]float64, 0, len(results))
	for _, r := range results {
		scores = append(scores, r.Score)
	}
	return scores, nil
}
