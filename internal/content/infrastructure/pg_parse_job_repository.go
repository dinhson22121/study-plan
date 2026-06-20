package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/content/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgParseJobRepository implements domain.ParseJobRepository over Postgres.
type PgParseJobRepository struct {
	db *pgxpool.Pool
}

// NewPgParseJobRepository builds the repository.
func NewPgParseJobRepository(db *pgxpool.Pool) *PgParseJobRepository {
	return &PgParseJobRepository{db: db}
}

func (r *PgParseJobRepository) Create(ctx context.Context, j *domain.ParseJob) error {
	const q = `
		INSERT INTO parse_job (id, asset_id, status, attempt_count, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.Exec(ctx, q, j.ID, j.AssetID, string(j.Status), j.AttemptCount, j.CreatedBy, j.CreatedAt, j.UpdatedAt)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgParseJobRepository) ListByAsset(ctx context.Context, assetID string) ([]domain.ParseJob, error) {
	const q = `
		SELECT id, asset_id, status, COALESCE(parser_version,''), attempt_count,
		       COALESCE(error_message,''), COALESCE(claimed_by,''), claimed_at, started_at, finished_at,
		       created_by, created_at, updated_at
		FROM parse_job WHERE asset_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, assetID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.ParseJob
	for rows.Next() {
		var j domain.ParseJob
		var status string
		if err := rows.Scan(&j.ID, &j.AssetID, &status, &j.ParserVersion, &j.AttemptCount,
			&j.ErrorMessage, &j.ClaimedBy, &j.ClaimedAt, &j.StartedAt, &j.FinishedAt,
			&j.CreatedBy, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		j.Status = domain.ParseJobStatus(status)
		out = append(out, j)
	}
	return out, rows.Err()
}

var _ domain.ParseJobRepository = (*PgParseJobRepository)(nil)
