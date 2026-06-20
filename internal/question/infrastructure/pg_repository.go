// Package infrastructure provides the Postgres adapter for the question
// repository.
package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const pgForeignKeyViolation = "23503"

// PgRepository implements domain.Repository over Postgres.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository builds the repository.
func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

// Create inserts a question and its options atomically. A reference to a missing
// topic surfaces as a validation error (not an internal error).
func (r *PgRepository) Create(ctx context.Context, q *domain.Question) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertQ = `
		INSERT INTO question (id, topic_id, type, stem, difficulty, explanation)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.Exec(ctx, insertQ, q.ID, q.TopicID, string(q.Type), q.Stem, string(q.Difficulty), q.Explanation); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return shared.ErrValidation.WithMessage("topic does not exist")
		}
		return shared.ErrInternal.WithCause(err)
	}

	const insertO = `
		INSERT INTO answer_option (id, question_id, text, is_correct, order_index)
		VALUES ($1, $2, $3, $4, $5)`
	for _, o := range q.Options {
		if _, err := tx.Exec(ctx, insertO, o.ID, q.ID, o.Text, o.IsCorrect, o.OrderIndex); err != nil {
			return shared.ErrInternal.WithCause(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// GetByID returns a question with its options ordered.
func (r *PgRepository) GetByID(ctx context.Context, id string) (*domain.Question, error) {
	const q = `SELECT id, topic_id, type, stem, difficulty, explanation FROM question WHERE id = $1`
	var qu domain.Question
	var qtype, difficulty string
	err := r.db.QueryRow(ctx, q, id).Scan(&qu.ID, &qu.TopicID, &qtype, &qu.Stem, &difficulty, &qu.Explanation)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("question not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	qu.Type = domain.QuestionType(qtype)
	qu.Difficulty = domain.Difficulty(difficulty)

	opts, err := r.loadOptions(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	qu.Options = opts[id]
	return &qu, nil
}

// List queries questions by filter and attaches their options.
func (r *PgRepository) List(ctx context.Context, f domain.ListFilter) ([]domain.Question, error) {
	const base = `SELECT id, topic_id, type, stem, difficulty, explanation FROM question WHERE topic_id = $1`
	args := []any{f.TopicID}
	query := base
	if f.Difficulty != "" {
		query += ` AND difficulty = $2`
		args = append(args, string(f.Difficulty))
	}
	query += ` ORDER BY created_at`
	if f.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, len(args)+1)
		args = append(args, f.Limit)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	var questions []domain.Question
	var ids []string
	for rows.Next() {
		var qu domain.Question
		var qtype, difficulty string
		if err := rows.Scan(&qu.ID, &qu.TopicID, &qtype, &qu.Stem, &difficulty, &qu.Explanation); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		qu.Type = domain.QuestionType(qtype)
		qu.Difficulty = domain.Difficulty(difficulty)
		questions = append(questions, qu)
		ids = append(ids, qu.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}

	if len(ids) == 0 {
		return questions, nil
	}
	opts, err := r.loadOptions(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range questions {
		questions[i].Options = opts[questions[i].ID]
	}
	return questions, nil
}

// loadOptions fetches options for many questions in one query (avoids N+1).
func (r *PgRepository) loadOptions(ctx context.Context, questionIDs []string) (map[string][]domain.AnswerOption, error) {
	const q = `
		SELECT id, question_id, text, is_correct, order_index
		FROM answer_option WHERE question_id = ANY($1) ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, questionIDs)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := map[string][]domain.AnswerOption{}
	for rows.Next() {
		var o domain.AnswerOption
		var qid string
		if err := rows.Scan(&o.ID, &qid, &o.Text, &o.IsCorrect, &o.OrderIndex); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out[qid] = append(out[qid], o)
	}
	return out, rows.Err()
}

var _ domain.Repository = (*PgRepository)(nil)
