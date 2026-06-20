package application

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
)

// Template codes seeded by migration 003 and used by the scheduler.
const (
	TemplateDailyReminder = "DAILY_REMINDER_V1"
	TemplateWeeklyQuiz    = "WEEKLY_QUIZ_V1"
)

// Scheduler runs the time-based notification jobs (PRD section 4). It is
// timezone-aware: cron entries fire in the configured location. Each job
// enqueues through the Dispatcher so the preference gate and idempotency still
// apply.
type Scheduler struct {
	cron       *cron.Cron
	dispatcher *Dispatcher
	repo       domain.Repository
	log        *zap.Logger
	now        func() time.Time
}

// NewScheduler builds a scheduler bound to the given timezone (falls back to UTC
// if the location name is invalid).
func NewScheduler(dispatcher *Dispatcher, repo domain.Repository, log *zap.Logger, timezone string) *Scheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Warn("invalid timezone, falling back to UTC", zap.String("timezone", timezone))
		loc = time.UTC
	}
	return &Scheduler{
		cron:       cron.New(cron.WithLocation(loc)),
		dispatcher: dispatcher,
		repo:       repo,
		log:        log,
		now:        func() time.Time { return time.Now().In(loc) },
	}
}

// Start registers the cron jobs and begins the scheduler. Call Stop to halt.
func (s *Scheduler) Start() error {
	jobs := []struct {
		spec string
		name string
		fn   func()
	}{
		{"0 20 * * *", "daily_reminder", s.runDailyReminder}, // every day 20:00
		{"0 19 * * 0", "weekly_quiz", s.runWeeklyQuiz},       // Sundays 19:00
		{"0 18 * * *", "reengagement", s.runReengagement},    // every day 18:00
	}
	for _, j := range jobs {
		if _, err := s.cron.AddFunc(j.spec, j.fn); err != nil {
			return fmt.Errorf("schedule %s: %w", j.name, err)
		}
	}
	s.cron.Start()
	s.log.Info("notification scheduler started")
	return nil
}

// Stop halts the scheduler, waiting for running jobs to finish.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}

func (s *Scheduler) runDailyReminder() {
	s.fanOut(domain.TypeDailyReminder, TemplateDailyReminder, "daily")
}

func (s *Scheduler) runWeeklyQuiz() {
	s.fanOut(domain.TypeWeeklyQuiz, TemplateWeeklyQuiz, "weekly")
}

// runReengagement is a placeholder until analytics (Phase 5) can identify
// inactive users and their weakest topic. It logs rather than sending blindly.
func (s *Scheduler) runReengagement() {
	s.log.Info("reengagement job skipped: pending analytics module (Phase 5)")
}

// fanOut enqueues a notification for every active user with a per-day
// idempotency key so re-runs within the same day do not duplicate.
func (s *Scheduler) fanOut(t domain.NotificationType, templateCode, cadence string) {
	ctx := context.Background()
	userIDs, err := s.repo.ListActiveUserIDs(ctx)
	if err != nil {
		s.log.Error("scheduler: list active users failed", zap.Error(err))
		return
	}
	day := s.now().Format("2006-01-02")
	sent := 0
	for _, uid := range userIDs {
		key := fmt.Sprintf("%s-%s-%s-%s", uid, t, cadence, day)
		if err := s.dispatcher.Enqueue(ctx, EnqueueInput{
			UserID:         uid,
			Type:           t,
			TemplateCode:   templateCode,
			IdempotencyKey: key,
		}); err != nil {
			s.log.Error("scheduler: enqueue failed", zap.String("user_id", uid), zap.Error(err))
			continue
		}
		sent++
	}
	s.log.Info("scheduler fan-out complete", zap.String("type", string(t)), zap.Int("enqueued", sent))
}
