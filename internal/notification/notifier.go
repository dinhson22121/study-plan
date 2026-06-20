package notification

import (
	"context"

	"github.com/son-ngo/edu-app/internal/notification/application"
	"github.com/son-ngo/edu-app/internal/notification/domain"
)

// notifier adapts the notification Dispatcher to the app.Notifier port so other
// modules can enqueue reminders without importing notification internals.
type notifier struct {
	dispatcher *application.Dispatcher
}

// EnqueueReminder enqueues a notification through the preference gate +
// idempotency + Kafka pipeline.
func (n *notifier) EnqueueReminder(ctx context.Context, userID, notifType, templateCode string, vars map[string]string, idempotencyKey string) error {
	return n.dispatcher.Enqueue(ctx, application.EnqueueInput{
		UserID:         userID,
		Type:           domain.NotificationType(notifType),
		TemplateCode:   templateCode,
		Variables:      vars,
		IdempotencyKey: idempotencyKey,
	})
}
