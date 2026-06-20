package notification

import (
	"context"

	"github.com/son-ngo/edu-app/internal/notification/application"
	"github.com/son-ngo/edu-app/internal/notification/domain"
)

type notifier struct {
	dispatcher *application.Dispatcher
}

func (n *notifier) EnqueueReminder(ctx context.Context, userID, notifType, templateCode string, vars map[string]string, idempotencyKey string) error {
	return n.dispatcher.Enqueue(ctx, application.EnqueueInput{
		UserID:         userID,
		Type:           domain.NotificationType(notifType),
		TemplateCode:   templateCode,
		Variables:      vars,
		IdempotencyKey: idempotencyKey,
	})
}
