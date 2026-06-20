package domain

import "context"

type Repository interface {
	GetTopicProgress(ctx context.Context, userID, topicID string) (*TopicProgress, error)
	UpsertTopicProgress(ctx context.Context, p *TopicProgress) error
	ListProgressByUser(ctx context.Context, userID string) ([]TopicProgress, error)

	GetStreak(ctx context.Context, userID string) (*Streak, error)
	UpsertStreak(ctx context.Context, s *Streak) error

	HasAchievement(ctx context.Context, userID string, t AchievementType, ref string) (bool, error)
	SaveAchievement(ctx context.Context, a *Achievement) error
}

type TopicTitleSource interface {
	Title(ctx context.Context, topicID string) (string, error)
}
