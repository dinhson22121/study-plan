package domain

import (
	"testing"
	"time"
)

func TestTopicProgress_RecordAttempt(t *testing.T) {
	p := TopicProgress{UserID: "u1", TopicID: "t1", Status: StatusNotStarted}
	now := time.Unix(1000, 0)

	p = p.RecordAttempt(60, now)
	if p.Status != StatusInProgress || p.BestScore != 60 || p.Attempts != 1 {
		t.Fatalf("after 60%%: %+v", p)
	}
	p = p.RecordAttempt(85, now)
	if p.Status != StatusCompleted || p.BestScore != 85 || p.Attempts != 2 {
		t.Fatalf("after 85%%: %+v", p)
	}

	p = p.RecordAttempt(40, now)
	if p.BestScore != 85 || p.Status != StatusCompleted {
		t.Fatalf("best score should not drop: %+v", p)
	}
}

func TestStreak_RecordActivity(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 1, d, 10, 0, 0, 0, time.UTC) }
	s := Streak{UserID: "u1"}

	s = s.RecordActivity(day(1))
	if s.CurrentStreak != 1 || s.LongestStreak != 1 {
		t.Fatalf("first day: %+v", s)
	}
	s = s.RecordActivity(day(1))
	if s.CurrentStreak != 1 {
		t.Fatalf("same day should not increment: %+v", s)
	}
	s = s.RecordActivity(day(2))
	if s.CurrentStreak != 2 {
		t.Fatalf("consecutive day: %+v", s)
	}
	s = s.RecordActivity(day(5))
	if s.CurrentStreak != 1 || s.LongestStreak != 2 {
		t.Fatalf("gap should reset, keep longest: %+v", s)
	}
}

func TestDetectAchievements(t *testing.T) {
	now := time.Unix(0, 0)

	got := DetectAchievements("u1", "t1", 100, true, 3, 4, now)
	if len(got) != 2 {
		t.Fatalf("expected TOPIC_COMPLETED + PERFECT_SCORE, got %d: %+v", len(got), got)
	}

	got = DetectAchievements("u1", "t1", 50, false, 6, 7, now)
	if len(got) != 1 || got[0].Type != AchievementStreak7 {
		t.Fatalf("expected STREAK_7, got %+v", got)
	}

	got = DetectAchievements("u1", "t1", 50, false, 7, 8, now)
	if len(got) != 0 {
		t.Fatalf("expected none, got %+v", got)
	}

	got = DetectAchievements("u1", "t1", 50, false, 29, 30, now)
	if len(got) != 1 || got[0].Type != AchievementStreak30 {
		t.Fatalf("expected STREAK_30, got %+v", got)
	}
}
