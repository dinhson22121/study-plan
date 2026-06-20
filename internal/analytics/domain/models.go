// Package domain defines the analytics bounded context: read-model shapes for
// dashboards and the inactivity feed, plus the ports it reads through.
package domain

import (
	"context"
	"time"
)

// TopicStat is a per-topic mastery snapshot used for weak-topic ranking.
type TopicStat struct {
	TopicID   string
	Completed bool
	BestScore float64
}

// ProgressSnapshot is the progress data analytics needs for the dashboard.
type ProgressSnapshot struct {
	CurrentStreak   int
	LongestStreak   int
	TopicsTotal     int
	TopicsCompleted int
	Topics          []TopicStat
}

// Dashboard is the aggregated learner dashboard.
type Dashboard struct {
	CurrentStreak   int     `json:"current_streak"`
	LongestStreak   int     `json:"longest_streak"`
	TopicsCompleted int     `json:"topics_completed"`
	TopicsTotal     int     `json:"topics_total"`
	QuizAverage     float64 `json:"quiz_average"`
	QuizCount       int     `json:"quiz_count"`
}

// WeakTopic is a topic the learner is weakest at (lowest best score, not mastered).
type WeakTopic struct {
	TopicID   string  `json:"topic_id"`
	BestScore float64 `json:"best_score"`
}

// ActivityRepo records and queries user activity for inactivity detection.
type ActivityRepo interface {
	Append(ctx context.Context, userID string, at time.Time) error
	// InactiveUserIDs returns users whose most recent activity is before the
	// given cutoff.
	InactiveUserIDs(ctx context.Context, before time.Time) ([]string, error)
}

// ProgressReader reads a user's progress snapshot (adapter over the progress module).
type ProgressReader interface {
	Snapshot(ctx context.Context, userID string) (ProgressSnapshot, error)
}

// QuizReader reads a user's quiz result scores (adapter over the quiz module).
type QuizReader interface {
	Scores(ctx context.Context, userID string) ([]float64, error)
}
