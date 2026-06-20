package domain

import (
	"context"
	"time"
)

// Repository persists all notification state: device tokens, templates,
// preferences, and the delivery log.
type Repository interface {
	// Device tokens
	UpsertDeviceToken(ctx context.Context, dt *DeviceToken) error
	FindActiveDeviceToken(ctx context.Context, userID string) (string, error)
	DeactivateToken(ctx context.Context, token string) error
	DeleteDeviceToken(ctx context.Context, userID, token string) error

	// Templates
	FindTemplate(ctx context.Context, code string) (*NotificationTemplate, error)

	// Preferences
	FindPreference(ctx context.Context, userID string, t NotificationType) (*NotificationPreference, error)
	ListPreferences(ctx context.Context, userID string) ([]NotificationPreference, error)
	UpsertPreference(ctx context.Context, p *NotificationPreference) error

	// Delivery log
	SaveLog(ctx context.Context, l *NotificationLog) error
	UpdateLogStatus(ctx context.Context, id string, status NotificationStatus, sentAt *time.Time, errMsg string) error
	ListLogsByUser(ctx context.Context, userID string, limit, offset int) ([]NotificationLog, int, error)

	// Audience
	ListActiveUserIDs(ctx context.Context) ([]string, error)
}

// FCMPort sends push notifications. Send encapsulates retry/backoff and returns
// a classified domain error (ErrTokenInvalid, ErrMaxRetriesExceeded) so callers
// can decide whether to retry or deactivate the token.
type FCMPort interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
}

// Publisher publishes a raw message to a Kafka topic. pkg/kafka.Producer
// satisfies this directly, keeping the domain free of Kafka imports.
type Publisher interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
}

// IdempotencyStore deduplicates schedule messages. CheckAndSet atomically
// records a key and reports whether this was the first time it was seen.
type IdempotencyStore interface {
	CheckAndSet(ctx context.Context, key string, ttl time.Duration) (firstSeen bool, err error)
}
