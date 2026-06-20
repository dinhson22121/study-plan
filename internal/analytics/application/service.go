// Package application contains the analytics use cases: building the learner
// dashboard, ranking weak topics, recording activity, and listing inactive
// users for re-engagement.
package application

import (
	"context"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/analytics/domain"
	quizdomain "github.com/son-ngo/edu-app/internal/quiz/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// Service implements the analytics use cases.
type Service struct {
	activity domain.ActivityRepo
	progress domain.ProgressReader
	quiz     domain.QuizReader
	log      *zap.Logger
	now      func() time.Time
}

// NewService builds the service.
func NewService(activity domain.ActivityRepo, progress domain.ProgressReader, quiz domain.QuizReader, log *zap.Logger) *Service {
	return &Service{activity: activity, progress: progress, quiz: quiz, log: log, now: time.Now}
}

// HandleQuizCompleted records activity from a quiz completion (best-effort).
func (s *Service) HandleQuizCompleted(ctx context.Context, evt shared.DomainEvent) error {
	e, ok := evt.(quizdomain.QuizCompletedEvent)
	if !ok {
		return nil
	}
	if err := s.activity.Append(ctx, e.UserID, s.now()); err != nil {
		s.log.Error("analytics: failed to record activity", zap.String("user_id", e.UserID), zap.Error(err))
	}
	return nil
}

// Dashboard aggregates progress + quiz data into the learner dashboard.
func (s *Service) Dashboard(ctx context.Context, userID string) (*domain.Dashboard, error) {
	ps, err := s.progress.Snapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	scores, err := s.quiz.Scores(ctx, userID)
	if err != nil {
		return nil, err
	}
	avg := 0.0
	if len(scores) > 0 {
		sum := 0.0
		for _, v := range scores {
			sum += v
		}
		avg = sum / float64(len(scores))
	}
	return &domain.Dashboard{
		CurrentStreak:   ps.CurrentStreak,
		LongestStreak:   ps.LongestStreak,
		TopicsCompleted: ps.TopicsCompleted,
		TopicsTotal:     ps.TopicsTotal,
		QuizAverage:     avg,
		QuizCount:       len(scores),
	}, nil
}

// WeakTopics returns up to n not-yet-mastered topics with the lowest best score.
func (s *Service) WeakTopics(ctx context.Context, userID string, n int) ([]domain.WeakTopic, error) {
	ps, err := s.progress.Snapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	var weak []domain.WeakTopic
	for _, t := range ps.Topics {
		if !t.Completed {
			weak = append(weak, domain.WeakTopic{TopicID: t.TopicID, BestScore: t.BestScore})
		}
	}
	sort.Slice(weak, func(i, j int) bool { return weak[i].BestScore < weak[j].BestScore })
	if n > 0 && len(weak) > n {
		weak = weak[:n]
	}
	return weak, nil
}

// InactiveUserIDs lists users inactive for at least the given number of days.
// Satisfies app.ReengagementSource for the notification re-engagement scheduler.
func (s *Service) InactiveUserIDs(ctx context.Context, days int) ([]string, error) {
	cutoff := s.now().AddDate(0, 0, -days)
	return s.activity.InactiveUserIDs(ctx, cutoff)
}
