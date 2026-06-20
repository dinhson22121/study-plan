package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgRepository implements domain.Repository over Postgres.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository builds the repository.
func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

// --- Device tokens ---

func (r *PgRepository) UpsertDeviceToken(ctx context.Context, dt *domain.DeviceToken) error {
	const q = `
		INSERT INTO device_token (id, user_id, device_token, platform, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, $5, $5)
		ON CONFLICT (device_token) DO UPDATE
		SET user_id = EXCLUDED.user_id, platform = EXCLUDED.platform,
		    is_active = true, updated_at = EXCLUDED.updated_at`
	_, err := r.db.Exec(ctx, q, dt.ID, dt.UserID, dt.Token, string(dt.Platform), dt.CreatedAt)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) FindActiveDeviceToken(ctx context.Context, userID string) (string, error) {
	const q = `
		SELECT device_token FROM device_token
		WHERE user_id = $1 AND is_active = true
		ORDER BY updated_at DESC LIMIT 1`
	var token string
	err := r.db.QueryRow(ctx, q, userID).Scan(&token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", shared.ErrNotFound
		}
		return "", shared.ErrInternal.WithCause(err)
	}
	return token, nil
}

func (r *PgRepository) DeactivateToken(ctx context.Context, token string) error {
	const q = `UPDATE device_token SET is_active = false, updated_at = NOW() WHERE device_token = $1`
	if _, err := r.db.Exec(ctx, q, token); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) DeleteDeviceToken(ctx context.Context, userID, token string) error {
	const q = `DELETE FROM device_token WHERE user_id = $1 AND device_token = $2`
	if _, err := r.db.Exec(ctx, q, userID, token); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// --- Templates ---

func (r *PgRepository) FindTemplate(ctx context.Context, code string) (*domain.NotificationTemplate, error) {
	const q = `SELECT code, title, body, notification_type, is_active FROM notification_template WHERE code = $1 AND is_active = true`
	var t domain.NotificationTemplate
	var notifType string
	err := r.db.QueryRow(ctx, q, code).Scan(&t.Code, &t.Title, &t.Body, &notifType, &t.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	t.Type = domain.NotificationType(notifType)
	return &t, nil
}

// --- Preferences ---

func (r *PgRepository) FindPreference(ctx context.Context, userID string, nt domain.NotificationType) (*domain.NotificationPreference, error) {
	const q = `SELECT enabled FROM notification_preference WHERE user_id = $1 AND notification_type = $2`
	var enabled bool
	err := r.db.QueryRow(ctx, q, userID, string(nt)).Scan(&enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &domain.NotificationPreference{UserID: userID, Type: nt, Enabled: enabled}, nil
}

func (r *PgRepository) ListPreferences(ctx context.Context, userID string) ([]domain.NotificationPreference, error) {
	const q = `SELECT notification_type, enabled FROM notification_preference WHERE user_id = $1`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	var out []domain.NotificationPreference
	for rows.Next() {
		var nt string
		var enabled bool
		if err := rows.Scan(&nt, &enabled); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out = append(out, domain.NotificationPreference{UserID: userID, Type: domain.NotificationType(nt), Enabled: enabled})
	}
	return out, rows.Err()
}

func (r *PgRepository) UpsertPreference(ctx context.Context, p *domain.NotificationPreference) error {
	const q = `
		INSERT INTO notification_preference (user_id, notification_type, enabled, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, notification_type) DO UPDATE
		SET enabled = EXCLUDED.enabled, updated_at = NOW()`
	if _, err := r.db.Exec(ctx, q, p.UserID, string(p.Type), p.Enabled); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// --- Delivery log ---

func (r *PgRepository) SaveLog(ctx context.Context, l *domain.NotificationLog) error {
	const q = `
		INSERT INTO notification_log
			(id, user_id, template_code, notification_type, correlation_id, status, retry_count, sent_at, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, q,
		l.ID, l.UserID, l.TemplateCode, string(l.Type), l.CorrelationID,
		string(l.Status), l.RetryCount, l.SentAt, l.ErrorMessage, l.CreatedAt)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) UpdateLogStatus(ctx context.Context, id string, status domain.NotificationStatus, sentAt *time.Time, errMsg string) error {
	const q = `UPDATE notification_log SET status = $2, sent_at = $3, error_message = $4 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, string(status), sentAt, errMsg)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *PgRepository) ListLogsByUser(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationLog, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notification_log WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, shared.ErrInternal.WithCause(err)
	}

	const q = `
		SELECT id, user_id, template_code, notification_type, correlation_id, status, retry_count, sent_at, error_message, created_at
		FROM notification_log WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	var out []domain.NotificationLog
	for rows.Next() {
		var l domain.NotificationLog
		var status, notifType string
		if err := rows.Scan(&l.ID, &l.UserID, &l.TemplateCode, &notifType, &l.CorrelationID,
			&status, &l.RetryCount, &l.SentAt, &l.ErrorMessage, &l.CreatedAt); err != nil {
			return nil, 0, shared.ErrInternal.WithCause(err)
		}
		l.Status = domain.NotificationStatus(status)
		l.Type = domain.NotificationType(notifType)
		out = append(out, l)
	}
	return out, total, rows.Err()
}

// --- Audience ---

func (r *PgRepository) ListActiveUserIDs(ctx context.Context) ([]string, error) {
	const q = `SELECT DISTINCT user_id FROM device_token WHERE is_active = true`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// compile-time check that PgRepository satisfies the port.
var _ domain.Repository = (*PgRepository)(nil)
