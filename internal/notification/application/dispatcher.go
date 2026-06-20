// Package application contains the notification use cases: dispatching
// (preference gate + idempotency + enqueue), the pipeline processors for each
// Kafka stage, preference/device-token management, and the cron scheduler.
package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// idempotencyTTL matches the PRD: dedupe a schedule key for 24h.
const idempotencyTTL = 24 * time.Hour

// Dispatcher implements the trigger side of the pipeline: it gates on user
// preference, deduplicates via the idempotency store, and enqueues a schedule
// message. Upstream modules and the cron scheduler both call Enqueue.
type Dispatcher struct {
	repo  domain.Repository
	idem  domain.IdempotencyStore
	pub   domain.Publisher
	log   *zap.Logger
	now   func() time.Time
	newID func() string
}

// NewDispatcher builds the dispatcher.
func NewDispatcher(repo domain.Repository, idem domain.IdempotencyStore, pub domain.Publisher, log *zap.Logger) *Dispatcher {
	return &Dispatcher{repo: repo, idem: idem, pub: pub, log: log, now: time.Now, newID: uuid.NewString}
}

// EnqueueInput is the dispatch command.
type EnqueueInput struct {
	UserID         string
	Type           domain.NotificationType
	TemplateCode   string
	Variables      map[string]string
	IdempotencyKey string // optional; generated when empty
	DeepLink       string
	CorrelationID  string // optional; generated when empty
}

// Enqueue runs the preference gate and idempotency check, then produces a
// schedule message. Suppressed (preference off) and duplicate (already enqueued)
// notifications are handled silently and return nil — they are expected outcomes,
// not errors.
func (d *Dispatcher) Enqueue(ctx context.Context, in EnqueueInput) error {
	if !in.Type.Valid() {
		return shared.ErrValidation.WithMessage("invalid notification type")
	}

	enabled, err := d.preferenceEnabled(ctx, in.UserID, in.Type)
	if err != nil {
		return err
	}
	if !enabled {
		// Preference gate: log SKIPPED and stop (PRD section 4 / step 2).
		skipped := domain.NewSkippedLog(d.newID(), in.UserID, in.TemplateCode, in.Type, d.correlationID(in), d.now())
		if err := d.repo.SaveLog(ctx, skipped); err != nil {
			return err
		}
		return nil
	}

	key := in.IdempotencyKey
	if key == "" {
		key = d.newID()
	}
	firstSeen, err := d.idem.CheckAndSet(ctx, key, idempotencyTTL)
	if err != nil {
		return err
	}
	if !firstSeen {
		d.log.Debug("notification deduplicated", zap.String("idempotency_key", key))
		return nil
	}

	msg := domain.ScheduleMessage{
		CorrelationID:     d.correlationID(in),
		StudentID:         in.UserID,
		NotificationType:  in.Type,
		TemplateCode:      in.TemplateCode,
		TemplateVariables: in.Variables,
		ScheduledAt:       d.now().Format(time.RFC3339),
		IdempotencyKey:    key,
		DeepLink:          in.DeepLink,
	}
	return d.publish(ctx, domain.TopicSchedule, in.UserID, msg)
}

func (d *Dispatcher) preferenceEnabled(ctx context.Context, userID string, t domain.NotificationType) (bool, error) {
	pref, err := d.repo.FindPreference(ctx, userID, t)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return true, nil // default: enabled when no explicit preference
		}
		return false, err
	}
	return pref.Enabled, nil
}

func (d *Dispatcher) correlationID(in EnqueueInput) string {
	if in.CorrelationID != "" {
		return in.CorrelationID
	}
	return d.newID()
}

func (d *Dispatcher) publish(ctx context.Context, topic, key string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if err := d.pub.Publish(ctx, topic, []byte(key), b); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}
