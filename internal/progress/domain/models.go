package domain

import "time"

const MasteryThreshold = 80.0

type ProgressStatus string

const (
	StatusNotStarted ProgressStatus = "NOT_STARTED"
	StatusInProgress ProgressStatus = "IN_PROGRESS"
	StatusCompleted  ProgressStatus = "COMPLETED"
)

type TopicProgress struct {
	UserID    string
	TopicID   string
	Status    ProgressStatus
	BestScore float64
	Attempts  int
	UpdatedAt time.Time
}

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

type Streak struct {
	UserID         string
	CurrentStreak  int
	LongestStreak  int
	LastActiveDate time.Time
}

func (s Streak) RecordActivity(now time.Time) Streak {
	today := dateOnly(now)
	switch {
	case s.LastActiveDate.IsZero():
		s.CurrentStreak = 1
	case today.Equal(dateOnly(s.LastActiveDate)):

	case today.Equal(dateOnly(s.LastActiveDate).AddDate(0, 0, 1)):
		s.CurrentStreak++
	case today.After(dateOnly(s.LastActiveDate)):
		s.CurrentStreak = 1
	default:
		return s
	}
	if s.CurrentStreak > s.LongestStreak {
		s.LongestStreak = s.CurrentStreak
	}
	s.LastActiveDate = today
	return s
}

func dateOnly(t time.Time) time.Time {
	t = t.UTC()
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
