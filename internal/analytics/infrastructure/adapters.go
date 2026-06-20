package infrastructure

import (
	"context"

	"github.com/son-ngo/edu-app/internal/analytics/domain"
	progressapp "github.com/son-ngo/edu-app/internal/progress/application"
	progressdomain "github.com/son-ngo/edu-app/internal/progress/domain"
	quizapp "github.com/son-ngo/edu-app/internal/quiz/application"
)

type ProgressReaderAdapter struct{ progress *progressapp.Service }

func NewProgressReaderAdapter(p *progressapp.Service) *ProgressReaderAdapter {
	return &ProgressReaderAdapter{progress: p}
}

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

type QuizReaderAdapter struct{ quiz *quizapp.Service }

func NewQuizReaderAdapter(q *quizapp.Service) *QuizReaderAdapter { return &QuizReaderAdapter{quiz: q} }

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
