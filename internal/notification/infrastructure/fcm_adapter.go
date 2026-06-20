// Package infrastructure provides the concrete adapters for the notification
// ports: the FCM sender (with retry/backoff), the Postgres repository, the
// Redis idempotency store, and the Kafka publisher.
package infrastructure

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// pushSender is the low-level FCM client capability the adapter wraps.
// pkg/fcm.Client satisfies it.
type pushSender interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
	IsTokenInvalid(err error) bool
}

// FCMAdapter implements domain.FCMPort. It retries transient failures with
// exponential backoff (1s, 2s, 4s) and, on an invalid-token error, deactivates
// the token immediately without retrying (PRD section 7).
type FCMAdapter struct {
	sender     pushSender
	repo       domain.Repository
	log        *zap.Logger
	maxRetries int
	wait       func(ctx context.Context, d time.Duration) error
}

// NewFCMAdapter wires the adapter to a real sender and repository.
func NewFCMAdapter(sender pushSender, repo domain.Repository, log *zap.Logger) *FCMAdapter {
	return &FCMAdapter{
		sender:     sender,
		repo:       repo,
		log:        log,
		maxRetries: 3,
		wait:       ctxSleep,
	}
}

// ctxSleep waits for d or until ctx is cancelled (returning ctx.Err()), so a
// shutdown mid-backoff is not blocked for the full delay.
func ctxSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// Send pushes a notification, retrying transient errors. It returns
// ErrTokenInvalid (after deactivating the token) for unregistered tokens, or
// ErrMaxRetriesExceeded after exhausting retries.
func (a *FCMAdapter) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	var lastErr error
	for attempt := 0; attempt < a.maxRetries; attempt++ {
		err := a.sender.Send(ctx, token, title, body, data)
		if err == nil {
			return nil
		}
		lastErr = err

		if a.sender.IsTokenInvalid(err) {
			if derr := a.repo.DeactivateToken(ctx, token); derr != nil {
				a.log.Warn("failed to deactivate invalid token", zap.Error(derr))
			}
			return shared.ErrTokenInvalid.WithCause(err)
		}

		// Exponential backoff before the next attempt: 1s, 2s, 4s.
		if attempt < a.maxRetries-1 {
			if werr := a.wait(ctx, time.Duration(1<<attempt)*time.Second); werr != nil {
				return shared.ErrMaxRetriesExceeded.WithCause(werr)
			}
		}
	}
	return shared.ErrMaxRetriesExceeded.WithCause(lastErr)
}
