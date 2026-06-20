// Package domain defines the progress bounded context: per-topic mastery,
// streaks, achievements, and the rules that evolve them from quiz activity.
package domain

import "time"

// MasteryThreshold is the best-score percent at which a topic is COMPLETED.
const MasteryThreshold = 80.0

// ProgressStatus is the mastery state of a topic for a user.
type ProgressStatus string

const (
	StatusNotStarted ProgressStatus = "NOT_STARTED"
	StatusInProgress ProgressStatus = "IN_PROGRESS"
	StatusCompleted  ProgressStatus = "COMPLETED"
)

// TopicProgress tracks a user's mastery of one topic.
type TopicProgress struct {
	UserID    string
	TopicID   string
	Status    ProgressStatus
	BestScore float64
	Attempts  int
	UpdatedAt time.Time
}

// RecordAttempt returns a new TopicProgress reflecting a quiz attempt (immutable
// update): attempts incremented, best score kept, status recomputed.
func (p TopicProgress) RecordAttempt(score float64, now time.Time) TopicProgress {
	p.Attempts++
	if score > p.BestScore {
		p.BestScore = score
	}
	if p.BestScore >= MasteryThreshold {
		p.Status = StatusCompleted
	} else {
		p.Status = StatusInProgress
	}
	p.UpdatedAt = now
	return p
}

// Streak tracks consecutive active days for a user.
type Streak struct {
	UserID         string
	CurrentStreak  int
	LongestStreak  int
	LastActiveDate time.Time
}

// RecordActivity returns a new Streak after activity on the given day. Same-day
// activity is a no-op; a consecutive day increments; any gap resets to 1.
func (s Streak) RecordActivity(now time.Time) Streak {
	today := dateOnly(now)
	switch {
	case s.LastActiveDate.IsZero():
		s.CurrentStreak = 1
	case today.Equal(dateOnly(s.LastActiveDate)):
		// already counted today
	case today.Equal(dateOnly(s.LastActiveDate).AddDate(0, 0, 1)):
		s.CurrentStreak++
	case today.After(dateOnly(s.LastActiveDate)):
		s.CurrentStreak = 1 // gap -> reset
	default:
		return s // activity in the past; ignore
	}
	if s.CurrentStreak > s.LongestStreak {
		s.LongestStreak = s.CurrentStreak
	}
	s.LastActiveDate = today
	return s
}

// dateOnly truncates to the calendar day in UTC. Normalizing to UTC keeps streak
// comparisons stable regardless of the server timezone vs the timestamptz values
// loaded from Postgres (which arrive as UTC).
func dateOnly(t time.Time) time.Time {
	t = t.UTC()
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
