// Package application contains the progress use cases: reacting to quiz
// completion (mastery, streaks, achievements) and reading a user's progress.
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/progress/domain"
	quizdomain "github.com/son-ngo/edu-app/internal/quiz/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// Service implements the progress use cases.
type Service struct {
	repo     domain.Repository
	titles   domain.TopicTitleSource
	notifier app.Notifier
	log      *zap.Logger
	now      func() time.Time
}

// NewService builds the service.
func NewService(repo domain.Repository, titles domain.TopicTitleSource, notifier app.Notifier, log *zap.Logger) *Service {
	return &Service{repo: repo, titles: titles, notifier: notifier, log: log, now: time.Now}
}

// HandleQuizCompleted reacts to a quiz.completed event: it updates topic mastery
// and the streak, then awards any newly-earned achievements (with a push). It is
// best-effort — errors are logged and swallowed so a failure here never fails the
// quiz submission that published the event (the bus is synchronous).
func (s *Service) HandleQuizCompleted(ctx context.Context, evt shared.DomainEvent) error {
	e, ok := evt.(quizdomain.QuizCompletedEvent)
	if !ok {
		s.log.Warn("progress: unexpected event type", zap.String("event", evt.EventName()))
		return nil
	}
	if err := s.apply(ctx, e); err != nil {
		s.log.Error("progress: failed to apply quiz.completed",
			zap.String("user_id", e.UserID), zap.String("topic_id", e.TopicID), zap.Error(err))
	}
	return nil
}

func (s *Service) apply(ctx context.Context, e quizdomain.QuizCompletedEvent) error {
	now := s.now()

	// Mastery.
	prog, err := s.loadProgress(ctx, e.UserID, e.TopicID)
	if err != nil {
		return err
	}
	wasCompleted := prog.Status == domain.StatusCompleted
	newProg := prog.RecordAttempt(e.Score, now)
	if err := s.repo.UpsertTopicProgress(ctx, &newProg); err != nil {
		return err
	}
	justCompleted := newProg.Status == domain.StatusCompleted && !wasCompleted

	// Streak.
	streak, err := s.loadStreak(ctx, e.UserID)
	if err != nil {
		return err
	}
	streakBefore := streak.CurrentStreak
	newStreak := streak.RecordActivity(now)
	if err := s.repo.UpsertStreak(ctx, &newStreak); err != nil {
		return err
	}

	// Achievements.
	earned := domain.DetectAchievements(e.UserID, e.TopicID, e.Score, justCompleted, streakBefore, newStreak.CurrentStreak, now)
	for i := range earned {
		if err := s.award(ctx, earned[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) loadProgress(ctx context.Context, userID, topicID string) (domain.TopicProgress, error) {
	prog, err := s.repo.GetTopicProgress(ctx, userID, topicID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return domain.TopicProgress{UserID: userID, TopicID: topicID, Status: domain.StatusNotStarted}, nil
		}
		return domain.TopicProgress{}, err
	}
	return *prog, nil
}

func (s *Service) loadStreak(ctx context.Context, userID string) (domain.Streak, error) {
	streak, err := s.repo.GetStreak(ctx, userID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return domain.Streak{UserID: userID}, nil
		}
		return domain.Streak{}, err
	}
	return *streak, nil
}

// award records a new achievement (once) and sends its push.
func (s *Service) award(ctx context.Context, a domain.Achievement) error {
	has, err := s.repo.HasAchievement(ctx, a.UserID, a.Type, a.Ref)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	if err := s.repo.SaveAchievement(ctx, &a); err != nil {
		return err
	}
	s.push(ctx, a)
	return nil
}

// push sends the achievement notification (best-effort).
func (s *Service) push(ctx context.Context, a domain.Achievement) {
	if s.notifier == nil {
		return
	}
	label := s.achievementLabel(ctx, a)
	idemKey := fmt.Sprintf("%s:%s:%s", a.UserID, a.Type, a.Ref)
	if err := s.notifier.EnqueueReminder(ctx, a.UserID, "ACHIEVEMENT", "ACHIEVEMENT_V1",
		map[string]string{"topic": label}, idemKey); err != nil {
		s.log.Warn("progress: achievement push failed", zap.String("user_id", a.UserID), zap.Error(err))
	}
}

// achievementLabel produces the {topic} variable for the push copy.
func (s *Service) achievementLabel(ctx context.Context, a domain.Achievement) string {
	switch a.Type {
	case domain.AchievementStreak7:
		return "chuỗi 7 ngày học liên tục"
	case domain.AchievementStreak30:
		return "chuỗi 30 ngày học liên tục"
	default: // TOPIC_COMPLETED, PERFECT_SCORE -> ref is a topic id
		if title, err := s.titles.Title(ctx, a.Ref); err == nil && title != "" {
			return title
		}
		return "một chủ đề"
	}
}

// --- queries ---

// Overview is the progress summary for a user.
type Overview struct {
	CurrentStreak int                    `json:"current_streak"`
	LongestStreak int                    `json:"longest_streak"`
	TopicsTotal   int                    `json:"topics_total"`
	TopicsDone    int                    `json:"topics_completed"`
	Topics        []domain.TopicProgress `json:"topics"`
}

// GetOverview returns a user's streak and per-topic progress.
func (s *Service) GetOverview(ctx context.Context, userID string) (*Overview, error) {
	topics, err := s.repo.ListProgressByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	streak, err := s.loadStreak(ctx, userID)
	if err != nil {
		return nil, err
	}
	done := 0
	for _, t := range topics {
		if t.Status == domain.StatusCompleted {
			done++
		}
	}
	return &Overview{
		CurrentStreak: streak.CurrentStreak, LongestStreak: streak.LongestStreak,
		TopicsTotal: len(topics), TopicsDone: done, Topics: topics,
	}, nil
}
