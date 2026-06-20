package domain

import "time"

type AchievementType string

const (
	AchievementTopicCompleted AchievementType = "TOPIC_COMPLETED"
	AchievementStreak7        AchievementType = "STREAK_7"
	AchievementStreak30       AchievementType = "STREAK_30"
	AchievementPerfectScore   AchievementType = "PERFECT_SCORE"
)

const (
	streak7Days  = 7
	streak30Days = 30
	perfectScore = 100.0
)

type Achievement struct {
	UserID     string
	Type       AchievementType
	Ref        string
	UnlockedAt time.Time
}

func DetectAchievements(userID, topicID string, score float64, justCompleted bool, streakBefore, streakAfter int, now time.Time) []Achievement {
	var out []Achievement
	add := func(t AchievementType, ref string) {
		out = append(out, Achievement{UserID: userID, Type: t, Ref: ref, UnlockedAt: now})
	}

	if justCompleted {
		add(AchievementTopicCompleted, topicID)
	}
	if score >= perfectScore {
		add(AchievementPerfectScore, topicID)
	}
	if streakBefore < streak7Days && streakAfter >= streak7Days {
		add(AchievementStreak7, "7")
	}
	if streakBefore < streak30Days && streakAfter >= streak30Days {
		add(AchievementStreak30, "30")
	}
	return out
}
