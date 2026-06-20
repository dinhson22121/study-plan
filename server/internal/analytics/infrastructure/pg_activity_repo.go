package infrastructure

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type PgActivityRepo struct {
	db *pgxpool.Pool
}

func NewPgActivityRepo(db *pgxpool.Pool) *PgActivityRepo { return &PgActivityRepo{db: db} }

func (r *PgActivityRepo) Append(ctx context.Context, userID string, at time.Time) error {
	const q = `INSERT INTO activity_event (user_id, occurred_at) VALUES ($1, $2)`
	if _, err := r.db.Exec(ctx, q, userID, at); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

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
