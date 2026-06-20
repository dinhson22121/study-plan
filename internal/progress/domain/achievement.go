package domain

import "time"

// AchievementType enumerates the achievements that trigger a push.
type AchievementType string

const (
	AchievementTopicCompleted AchievementType = "TOPIC_COMPLETED"
	AchievementStreak7        AchievementType = "STREAK_7"
	AchievementStreak30       AchievementType = "STREAK_30"
	AchievementPerfectScore   AchievementType = "PERFECT_SCORE"
)

// Streak-milestone thresholds.
const (
	streak7Days  = 7
	streak30Days = 30
	perfectScore = 100.0
)

// Achievement is a unlocked badge, recorded once per (user, type, ref) so the
// push is not sent twice. Ref disambiguates (e.g. the topic id).
type Achievement struct {
	UserID     string
	Type       AchievementType
	Ref        string
	UnlockedAt time.Time
}

// DetectAchievements returns the achievements newly earned by a quiz attempt,
// given the score, whether the topic just reached COMPLETED, and the streak
// before/after the attempt. De-duplication against already-stored achievements
// is the caller's responsibility.
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
