package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/notification/domain"
)

// Manager implements the user-facing and admin use cases: device-token
// registration, preference reads/updates, delivery history, and broadcast.
type Manager struct {
	repo       domain.Repository
	dispatcher *Dispatcher
	now        func() time.Time
	newID      func() string
}

// NewManager builds the manager.
func NewManager(repo domain.Repository, dispatcher *Dispatcher) *Manager {
	return &Manager{repo: repo, dispatcher: dispatcher, now: time.Now, newID: uuid.NewString}
}

// RegisterDeviceToken upserts a user's FCM token (reactivating it if it existed).
func (m *Manager) RegisterDeviceToken(ctx context.Context, userID, token, platform string) error {
	dt, err := domain.NewDeviceToken(m.newID(), userID, token, domain.Platform(platform), m.now())
	if err != nil {
		return err
	}
	return m.repo.UpsertDeviceToken(ctx, dt)
}

// DeleteDeviceToken removes a token on logout.
func (m *Manager) DeleteDeviceToken(ctx context.Context, userID, token string) error {
	return m.repo.DeleteDeviceToken(ctx, userID, token)
}

// ListPreferences returns the user's notification preferences, defaulting to all
// types enabled when none are stored yet.
func (m *Manager) ListPreferences(ctx context.Context, userID string) ([]domain.NotificationPreference, error) {
	prefs, err := m.repo.ListPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(prefs) == 0 {
		return domain.DefaultPreferences(userID), nil
	}
	return prefs, nil
}

// SetPreference enables or disables one notification type for a user.
func (m *Manager) SetPreference(ctx context.Context, userID, notifType string, enabled bool) error {
	t, err := domain.ParseType(notifType)
	if err != nil {
		return err
	}
	return m.repo.UpsertPreference(ctx, &domain.NotificationPreference{UserID: userID, Type: t, Enabled: enabled})
}

// GetHistory returns a page of the user's notification log.
func (m *Manager) GetHistory(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationLog, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return m.repo.ListLogsByUser(ctx, userID, limit, offset)
}

// BroadcastInput is the admin broadcast command.
type BroadcastInput struct {
	Type         domain.NotificationType
	TemplateCode string
	Variables    map[string]string
}

// Broadcast fans an admin notification out to every active user, going through
// the dispatcher so each delivery still respects the user's preference gate.
// Returns the number of users enqueued.
func (m *Manager) Broadcast(ctx context.Context, in BroadcastInput) (int, error) {
	userIDs, err := m.repo.ListActiveUserIDs(ctx)
	if err != nil {
		return 0, err
	}
	// A deterministic per-day broadcast id gives each user a stable idempotency
	// key, so a retried or accidentally-repeated same-day broadcast of the same
	// template does not double-send.
	broadcastID := fmt.Sprintf("broadcast-%s-%s-%s", in.Type, in.TemplateCode, m.now().Format("2006-01-02"))
	count := 0
	for _, uid := range userIDs {
		err := m.dispatcher.Enqueue(ctx, EnqueueInput{
			UserID:         uid,
			Type:           in.Type,
			TemplateCode:   in.TemplateCode,
			Variables:      in.Variables,
			IdempotencyKey: broadcastID + ":" + uid,
		})
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
