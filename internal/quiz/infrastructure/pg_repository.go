package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/quiz/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

func (r *PgRepository) SaveSession(ctx context.Context, s *domain.QuizSession) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertS = `INSERT INTO quiz_session (id, user_id, topic_id, status, created_at) VALUES ($1,$2,$3,$4,$5)`
	if _, err := tx.Exec(ctx, insertS, s.ID, s.UserID, s.TopicID, string(s.Status), s.CreatedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	const insertQ = `INSERT INTO quiz_session_question (session_id, question_id, order_index) VALUES ($1,$2,$3)`
	for i, qid := range s.QuestionIDs {
		if _, err := tx.Exec(ctx, insertQ, s.ID, qid, i); err != nil {
			return shared.ErrInternal.WithCause(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) GetSession(ctx context.Context, id string) (*domain.QuizSession, error) {
	const q = `SELECT id, user_id, topic_id, status, created_at FROM quiz_session WHERE id = $1`
	var s domain.QuizSession
	var status string
	err := r.db.QueryRow(ctx, q, id).Scan(&s.ID, &s.UserID, &s.TopicID, &status, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("quiz session not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	s.Status = domain.TestStatus(status)

	rows, err := r.db.Query(ctx, `SELECT question_id FROM quiz_session_question WHERE session_id = $1 ORDER BY order_index`, id)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	for rows.Next() {
		var qid string
		if err := rows.Scan(&qid); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		s.QuestionIDs = append(s.QuestionIDs, qid)
	}
	return &s, rows.Err()
}

func (r *PgRepository) SaveResultAndComplete(ctx context.Context, res *domain.QuizResult) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertR = `
		INSERT INTO quiz_result (session_id, user_id, topic_id, score, correct_count, total, passed, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	if _, err := tx.Exec(ctx, insertR, res.SessionID, res.UserID, res.TopicID, res.Score,
		res.CorrectCount, res.Total, res.Passed, res.CompletedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	const insertA = `
		INSERT INTO quiz_answer (session_id, question_id, selected_option_id, is_correct)
		VALUES ($1,$2,$3,$4)`
	for _, rv := range res.Reviews {

		var selected any
		if rv.SelectedOptionID != "" {
			selected = rv.SelectedOptionID
		}
		if _, err := tx.Exec(ctx, insertA, res.SessionID, rv.QuestionID, selected, rv.IsCorrect); err != nil {
			return shared.ErrInternal.WithCause(err)
		}
	}
	tag, err := tx.Exec(ctx, `UPDATE quiz_session SET status = 'COMPLETED' WHERE id = $1`, res.SessionID)
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

func (r *PgRepository) GetResultForUser(ctx context.Context, sessionID, userID string) (*domain.QuizResult, error) {
	const q = `
		SELECT session_id, user_id, topic_id, score, correct_count, total, passed, completed_at
		FROM quiz_result WHERE session_id = $1 AND user_id = $2`
	res, err := scanResult(r.db.QueryRow(ctx, q, sessionID, userID))
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT question_id, selected_option_id, is_correct FROM quiz_answer WHERE session_id = $1`, sessionID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	for rows.Next() {
		var rv domain.QuestionReview
		var selected *string
		if err := rows.Scan(&rv.QuestionID, &selected, &rv.IsCorrect); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		if selected != nil {
			rv.SelectedOptionID = *selected
		}
		res.Reviews = append(res.Reviews, rv)
	}
	return res, rows.Err()
}

func (r *PgRepository) ListResultsByUser(ctx context.Context, userID string) ([]domain.QuizResult, error) {
	const q = `
		SELECT session_id, user_id, topic_id, score, correct_count, total, passed, completed_at
		FROM quiz_result WHERE user_id = $1 ORDER BY completed_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.QuizResult
	for rows.Next() {
		res, err := scanResult(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *res)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanResult(row rowScanner) (*domain.QuizResult, error) {
	var res domain.QuizResult
	err := row.Scan(&res.SessionID, &res.UserID, &res.TopicID, &res.Score,
		&res.CorrectCount, &res.Total, &res.Passed, &res.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("quiz result not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &res, nil
}

var _ domain.Repository = (*PgRepository)(nil)
