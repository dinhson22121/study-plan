package infrastructure

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgActivityRepo implements analytics's ActivityRepo over Postgres.
type PgActivityRepo struct {
	db *pgxpool.Pool
}

// NewPgActivityRepo builds the repository.
func NewPgActivityRepo(db *pgxpool.Pool) *PgActivityRepo { return &PgActivityRepo{db: db} }

// Append records an activity event for a user.
func (r *PgActivityRepo) Append(ctx context.Context, userID string, at time.Time) error {
	const q = `INSERT INTO activity_event (user_id, occurred_at) VALUES ($1, $2)`
	if _, err := r.db.Exec(ctx, q, userID, at); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// InactiveUserIDs returns users whose most recent activity is before cutoff.
func (r *PgActivityRepo) InactiveUserIDs(ctx context.Context, before time.Time) ([]string, error) {
	const q = `
		SELECT user_id FROM activity_event
		GROUP BY user_id
		HAVING MAX(occurred_at) < $1`
	rows, err := r.db.Query(ctx, q, before)
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
