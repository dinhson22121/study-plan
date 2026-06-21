package audit

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgRecorder persists audit entries to the admin_audit_log table (migration
// 016) via a pgx/v5 pool.
type PgRecorder struct {
	db *pgxpool.Pool
}

// NewPgRecorder returns a Postgres-backed Recorder.
func NewPgRecorder(db *pgxpool.Pool) *PgRecorder {
	return &PgRecorder{db: db}
}

func (r *PgRecorder) Record(ctx context.Context, entry Entry) error {
	const q = `
		INSERT INTO admin_audit_log
			(id, actor_user_id, action, method, path, status_code, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, q,
		uuid.New(),
		entry.ActorUserID,
		entry.Action(),
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.CorrelationID,
	)
	if err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}
