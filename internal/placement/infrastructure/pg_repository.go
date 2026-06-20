package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/placement/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgRepository implements domain.Repository over Postgres.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository builds the repository.
func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

// SaveTest inserts the test and its ordered question snapshot atomically.
func (r *PgRepository) SaveTest(ctx context.Context, t *domain.PlacementTest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertT = `INSERT INTO placement_test (id, user_id, subject_id, status, created_at) VALUES ($1,$2,$3,$4,$5)`
	if _, err := tx.Exec(ctx, insertT, t.ID, t.UserID, t.SubjectID, string(t.Status), t.CreatedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	const insertQ = `INSERT INTO placement_test_question (test_id, question_id, order_index) VALUES ($1,$2,$3)`
	for i, qid := range t.QuestionIDs {
		if _, err := tx.Exec(ctx, insertQ, t.ID, qid, i); err != nil {
			return shared.ErrInternal.WithCause(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// GetTest loads a test with its question snapshot.
func (r *PgRepository) GetTest(ctx context.Context, id string) (*domain.PlacementTest, error) {
	const q = `SELECT id, user_id, subject_id, status, created_at FROM placement_test WHERE id = $1`
	var t domain.PlacementTest
	var status string
	err := r.db.QueryRow(ctx, q, id).Scan(&t.ID, &t.UserID, &t.SubjectID, &status, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("placement test not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	t.Status = domain.TestStatus(status)

	rows, err := r.db.Query(ctx, `SELECT question_id FROM placement_test_question WHERE test_id = $1 ORDER BY order_index`, id)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	for rows.Next() {
		var qid string
		if err := rows.Scan(&qid); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		t.QuestionIDs = append(t.QuestionIDs, qid)
	}
	return &t, rows.Err()
}

// CompleteWithResult inserts the result and marks the test COMPLETED in one tx.
func (r *PgRepository) CompleteWithResult(ctx context.Context, testID string, res *domain.PlacementResult) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertResult = `INSERT INTO placement_result (id, user_id, subject_id, score, level, completed_at) VALUES ($1,$2,$3,$4,$5,$6)`
	if _, err := tx.Exec(ctx, insertResult, res.ID, res.UserID, res.SubjectID, res.Score, string(res.Level), res.CompletedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	tag, err := tx.Exec(ctx, `UPDATE placement_test SET status = 'COMPLETED' WHERE id = $1`, testID)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// ListResults returns a user's results, newest first.
func (r *PgRepository) ListResults(ctx context.Context, userID string) ([]domain.PlacementResult, error) {
	const q = `SELECT id, user_id, subject_id, score, level, completed_at FROM placement_result WHERE user_id = $1 ORDER BY completed_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.PlacementResult
	for rows.Next() {
		res, err := scanResult(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *res)
	}
	return out, rows.Err()
}

// LatestResult returns the most recent result for a user+subject.
func (r *PgRepository) LatestResult(ctx context.Context, userID, subjectID string) (*domain.PlacementResult, error) {
	const q = `
		SELECT id, user_id, subject_id, score, level, completed_at FROM placement_result
		WHERE user_id = $1 AND subject_id = $2 ORDER BY completed_at DESC LIMIT 1`
	rows, err := r.db.Query(ctx, q, userID, subjectID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, shared.ErrNotFound
	}
	return scanResult(rows)
}

func scanResult(rows pgx.Rows) (*domain.PlacementResult, error) {
	var res domain.PlacementResult
	var level string
	if err := rows.Scan(&res.ID, &res.UserID, &res.SubjectID, &res.Score, &level, &res.CompletedAt); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	res.Level = domain.Level(level)
	return &res, nil
}

var _ domain.Repository = (*PgRepository)(nil)
