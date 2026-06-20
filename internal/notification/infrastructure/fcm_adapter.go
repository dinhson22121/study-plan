package infrastructure

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type pushSender interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
	IsTokenInvalid(err error) bool
}

type FCMAdapter struct {
	sender     pushSender
	repo       domain.Repository
	log        *zap.Logger
	maxRetries int
	wait       func(ctx context.Context, d time.Duration) error
}

func NewFCMAdapter(sender pushSender, repo domain.Repository, log *zap.Logger) *FCMAdapter {
	return &FCMAdapter{
		sender:     sender,
		repo:       repo,
		log:        log,
		maxRetries: 3,
		wait:       ctxSleep,
	}
}

func ctxSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

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

		if attempt < a.maxRetries-1 {
			if werr := a.wait(ctx, time.Duration(1<<attempt)*time.Second); werr != nil {
				return shared.ErrMaxRetriesExceeded.WithCause(werr)
			}
		}
	}
	return shared.ErrMaxRetriesExceeded.WithCause(lastErr)
}
