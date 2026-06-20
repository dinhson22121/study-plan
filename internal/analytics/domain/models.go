package domain

import (
	"context"
	"time"
)

type TopicStat struct {
	TopicID   string
	Completed bool
	BestScore float64
}

type ProgressSnapshot struct {
	CurrentStreak   int
	LongestStreak   int
	TopicsTotal     int
	TopicsCompleted int
	Topics          []TopicStat
}

type Dashboard struct {
	CurrentStreak   int     `json:"current_streak"`
	LongestStreak   int     `json:"longest_streak"`
	TopicsCompleted int     `json:"topics_completed"`
	TopicsTotal     int     `json:"topics_total"`
	QuizAverage     float64 `json:"quiz_average"`
	QuizCount       int     `json:"quiz_count"`
}

type WeakTopic struct {
	TopicID   string  `json:"topic_id"`
	BestScore float64 `json:"best_score"`
}

type ActivityRepo interface {
	Append(ctx context.Context, userID string, at time.Time) error

	InactiveUserIDs(ctx context.Context, before time.Time) ([]string, error)
}

type ProgressReader interface {
	Snapshot(ctx context.Context, userID string) (ProgressSnapshot, error)
}

type QuizReader interface {
	Scores(ctx context.Context, userID string) ([]float64, error)
}
