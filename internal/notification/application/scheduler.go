package application

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
)

const (
	TemplateDailyReminder = "DAILY_REMINDER_V1"
	TemplateWeeklyQuiz    = "WEEKLY_QUIZ_V1"
	TemplateReengagement  = "REENGAGEMENT_V1"
)

const reengagementDays = 3

type ReengagementLister interface {
	InactiveUserIDs(ctx context.Context, days int) ([]string, error)
}

type Scheduler struct {
	cron         *cron.Cron
	dispatcher   *Dispatcher
	repo         domain.Repository
	reengagement ReengagementLister
	log          *zap.Logger
	now          func() time.Time
}

func NewScheduler(dispatcher *Dispatcher, repo domain.Repository, reengagement ReengagementLister, log *zap.Logger, timezone string) *Scheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Warn("invalid timezone, falling back to UTC", zap.String("timezone", timezone))
		loc = time.UTC
	}
	return &Scheduler{
		cron:         cron.New(cron.WithLocation(loc)),
		dispatcher:   dispatcher,
		repo:         repo,
		reengagement: reengagement,
		log:          log,
		now:          func() time.Time { return time.Now().In(loc) },
	}
}

func (s *Scheduler) Start() error {
	jobs := []struct {
		spec string
		name string
		fn   func()
	}{
		{"0 20 * * *", "daily_reminder", s.runDailyReminder},
		{"0 19 * * 0", "weekly_quiz", s.runWeeklyQuiz},
		{"0 18 * * *", "reengagement", s.runReengagement},
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

func (s *Scheduler) runReengagement() {
	if s.reengagement == nil {
		s.log.Info("reengagement job skipped: no reengagement source wired")
		return
	}
	ctx := context.Background()
	userIDs, err := s.reengagement.InactiveUserIDs(ctx, reengagementDays)
	if err != nil {
		s.log.Error("scheduler: list inactive users failed", zap.Error(err))
		return
	}
	day := s.now().Format("2006-01-02")
	sent := 0
	for _, uid := range userIDs {
		key := fmt.Sprintf("%s-reengage-%s", uid, day)
		if err := s.dispatcher.Enqueue(ctx, EnqueueInput{
			UserID:         uid,
			Type:           domain.TypeReengagement,
			TemplateCode:   TemplateReengagement,
			Variables:      map[string]string{"days": fmt.Sprint(reengagementDays)},
			IdempotencyKey: key,
		}); err != nil {
			s.log.Error("scheduler: reengagement enqueue failed", zap.String("user_id", uid), zap.Error(err))
			continue
		}
		sent++
	}
	s.log.Info("reengagement fan-out complete", zap.Int("enqueued", sent))
}

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
