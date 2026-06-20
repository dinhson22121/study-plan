package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgDraftRepository implements domain.DraftRepository over Postgres.
type PgDraftRepository struct {
	db *pgxpool.Pool
}

// NewPgDraftRepository builds the repository.
func NewPgDraftRepository(db *pgxpool.Pool) *PgDraftRepository { return &PgDraftRepository{db: db} }

const draftCols = `id, asset_id, parse_job_id, question_number, question_type, stem,
	COALESCE(explanation_raw,''), COALESCE(answer_key_raw,''), parse_confidence, status,
	COALESCE(reviewed_by,''), reviewed_at, COALESCE(published_question_id::text,''), created_at, updated_at`

func (r *PgDraftRepository) ListByAsset(ctx context.Context, assetID string) ([]domain.QuestionDraft, error) {
	rows, err := r.db.Query(ctx, `SELECT `+draftCols+` FROM question_draft WHERE asset_id = $1 ORDER BY question_number`, assetID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var drafts []domain.QuestionDraft
	var ids []string
	for rows.Next() {
		d, err := scanDraft(rows)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, *d)
		ids = append(ids, d.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	if len(ids) == 0 {
		return drafts, nil
	}
	opts, err := r.loadOptions(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range drafts {
		drafts[i].Options = opts[drafts[i].ID]
	}
	return drafts, nil
}

func (r *PgDraftRepository) GetByID(ctx context.Context, id string) (*domain.QuestionDraft, error) {
	d, err := scanDraft(r.db.QueryRow(ctx, `SELECT `+draftCols+` FROM question_draft WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	opts, err := r.loadOptions(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	d.Options = opts[id]
	return d, nil
}

func (r *PgDraftRepository) UpdateDraft(ctx context.Context, id, stem, explanation string, at time.Time) error {
	tag, err := r.db.Exec(ctx, `UPDATE question_draft SET stem = $2, explanation_raw = $3, updated_at = $4 WHERE id = $1`,
		id, stem, explanation, at)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *PgDraftRepository) UpdateOption(ctx context.Context, optionID, text string, isCorrect bool) error {
	tag, err := r.db.Exec(ctx, `UPDATE question_draft_option SET option_text = $2, is_correct_inferred = $3 WHERE id = $1`,
		optionID, text, isCorrect)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *PgDraftRepository) MarkPublished(ctx context.Context, id, questionID, reviewedBy string, at time.Time) error {
	const q = `
		UPDATE question_draft
		SET status = 'PUBLISHED', published_question_id = $2, reviewed_by = $3, reviewed_at = $4, updated_at = $4
		WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, questionID, reviewedBy, at)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *PgDraftRepository) loadOptions(ctx context.Context, draftIDs []string) (map[string][]domain.QuestionDraftOption, error) {
	const q = `
		SELECT id, question_draft_id, option_label, option_text, is_correct_inferred, order_index
		FROM question_draft_option WHERE question_draft_id = ANY($1) ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, draftIDs)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	out := map[string][]domain.QuestionDraftOption{}
	for rows.Next() {
		var o domain.QuestionDraftOption
		if err := rows.Scan(&o.ID, &o.QuestionDraftID, &o.OptionLabel, &o.OptionText, &o.IsCorrectInferred, &o.OrderIndex); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out[o.QuestionDraftID] = append(out[o.QuestionDraftID], o)
	}
	return out, rows.Err()
}

type rowScanner interface{ Scan(dest ...any) error }

func scanDraft(row rowScanner) (*domain.QuestionDraft, error) {
	var d domain.QuestionDraft
	var status string
	err := row.Scan(&d.ID, &d.AssetID, &d.ParseJobID, &d.QuestionNumber, &d.QuestionType, &d.Stem,
		&d.ExplanationRaw, &d.AnswerKeyRaw, &d.ParseConfidence, &status,
		&d.ReviewedBy, &d.ReviewedAt, &d.PublishedQuestionID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("question draft not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	d.Status = domain.DraftStatus(status)
	return &d, nil
}

var _ domain.DraftRepository = (*PgDraftRepository)(nil)
