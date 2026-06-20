package domain

import (
	"context"
	"time"
)

type Repository interface {
	UpsertDeviceToken(ctx context.Context, dt *DeviceToken) error
	FindActiveDeviceToken(ctx context.Context, userID string) (string, error)
	DeactivateToken(ctx context.Context, token string) error
	DeleteDeviceToken(ctx context.Context, userID, token string) error

	FindTemplate(ctx context.Context, code string) (*NotificationTemplate, error)

	FindPreference(ctx context.Context, userID string, t NotificationType) (*NotificationPreference, error)
	ListPreferences(ctx context.Context, userID string) ([]NotificationPreference, error)
	UpsertPreference(ctx context.Context, p *NotificationPreference) error

	SaveLog(ctx context.Context, l *NotificationLog) error
	UpdateLogStatus(ctx context.Context, id string, status NotificationStatus, sentAt *time.Time, errMsg string) error
	ListLogsByUser(ctx context.Context, userID string, limit, offset int) ([]NotificationLog, int, error)

	ListActiveUserIDs(ctx context.Context) ([]string, error)
}

type FCMPort interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
}

type Publisher interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
}

type IdempotencyStore interface {
	CheckAndSet(ctx context.Context, key string, ttl time.Duration) (firstSeen bool, err error)
}
