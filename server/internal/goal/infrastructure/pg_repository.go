package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/goal/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

func (r *PgRepository) Upsert(ctx context.Context, g *domain.Goal) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const upsertGoal = `
		INSERT INTO goal (user_id, target_university, target_major, target_date, hours_per_day, days_per_week, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			target_university = EXCLUDED.target_university,
			target_major = EXCLUDED.target_major,
			target_date = EXCLUDED.target_date,
			hours_per_day = EXCLUDED.hours_per_day,
			days_per_week = EXCLUDED.days_per_week,
			updated_at = EXCLUDED.updated_at`
	if _, err := tx.Exec(ctx, upsertGoal, g.UserID, g.TargetUniversity, g.TargetMajor, g.TargetDate,
		g.HoursPerDay, g.DaysPerWeek, g.UpdatedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM subject_target WHERE user_id = $1`, g.UserID); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	const insertTarget = `
		INSERT INTO subject_target (user_id, subject_id, current_score, target_score)
		VALUES ($1, $2, $3, $4)`
	for _, st := range g.Subjects {
		if _, err := tx.Exec(ctx, insertTarget, g.UserID, st.SubjectID, st.CurrentScore, st.TargetScore); err != nil {
			return shared.ErrInternal.WithCause(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) GetByUserID(ctx context.Context, userID string) (*domain.Goal, error) {
	const q = `
		SELECT user_id, target_university, target_major, target_date, hours_per_day, days_per_week, created_at, updated_at
		FROM goal WHERE user_id = $1`
	var g domain.Goal
	err := r.db.QueryRow(ctx, q, userID).Scan(&g.UserID, &g.TargetUniversity, &g.TargetMajor, &g.TargetDate,
		&g.HoursPerDay, &g.DaysPerWeek, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("goal not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}

	rows, err := r.db.Query(ctx, `SELECT subject_id, current_score, target_score FROM subject_target WHERE user_id = $1`, userID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	for rows.Next() {
		var st domain.SubjectTarget
		if err := rows.Scan(&st.SubjectID, &st.CurrentScore, &st.TargetScore); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		g.Subjects = append(g.Subjects, st)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &g, nil
}

var _ domain.Repository = (*PgRepository)(nil)
