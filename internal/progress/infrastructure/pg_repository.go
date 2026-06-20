package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/progress/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgRepository implements domain.Repository over Postgres.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository builds the repository.
func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

func (r *PgRepository) GetTopicProgress(ctx context.Context, userID, topicID string) (*domain.TopicProgress, error) {
	const q = `SELECT user_id, topic_id, status, best_score, attempts, updated_at FROM topic_progress WHERE user_id = $1 AND topic_id = $2`
	var p domain.TopicProgress
	var status string
	err := r.db.QueryRow(ctx, q, userID, topicID).Scan(&p.UserID, &p.TopicID, &status, &p.BestScore, &p.Attempts, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	p.Status = domain.ProgressStatus(status)
	return &p, nil
}

func (r *PgRepository) UpsertTopicProgress(ctx context.Context, p *domain.TopicProgress) error {
	const q = `
		INSERT INTO topic_progress (user_id, topic_id, status, best_score, attempts, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (user_id, topic_id) DO UPDATE SET
			status = EXCLUDED.status, best_score = EXCLUDED.best_score,
			attempts = EXCLUDED.attempts, updated_at = EXCLUDED.updated_at`
	if _, err := r.db.Exec(ctx, q, p.UserID, p.TopicID, string(p.Status), p.BestScore, p.Attempts, p.UpdatedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) ListProgressByUser(ctx context.Context, userID string) ([]domain.TopicProgress, error) {
	const q = `SELECT user_id, topic_id, status, best_score, attempts, updated_at FROM topic_progress WHERE user_id = $1 ORDER BY updated_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.TopicProgress
	for rows.Next() {
		var p domain.TopicProgress
		var status string
		if err := rows.Scan(&p.UserID, &p.TopicID, &status, &p.BestScore, &p.Attempts, &p.UpdatedAt); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		p.Status = domain.ProgressStatus(status)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PgRepository) GetStreak(ctx context.Context, userID string) (*domain.Streak, error) {
	const q = `SELECT user_id, current_streak, longest_streak, last_active_date FROM streak WHERE user_id = $1`
	var s domain.Streak
	err := r.db.QueryRow(ctx, q, userID).Scan(&s.UserID, &s.CurrentStreak, &s.LongestStreak, &s.LastActiveDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &s, nil
}

func (r *PgRepository) UpsertStreak(ctx context.Context, s *domain.Streak) error {
	const q = `
		INSERT INTO streak (user_id, current_streak, longest_streak, last_active_date)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id) DO UPDATE SET
			current_streak = EXCLUDED.current_streak, longest_streak = EXCLUDED.longest_streak,
			last_active_date = EXCLUDED.last_active_date`
	if _, err := r.db.Exec(ctx, q, s.UserID, s.CurrentStreak, s.LongestStreak, s.LastActiveDate); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) HasAchievement(ctx context.Context, userID string, t domain.AchievementType, ref string) (bool, error) {
	const q = `SELECT 1 FROM achievement WHERE user_id = $1 AND type = $2 AND ref = $3`
	var one int
	err := r.db.QueryRow(ctx, q, userID, string(t), ref).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, shared.ErrInternal.WithCause(err)
	}
	return true, nil
}

func (r *PgRepository) SaveAchievement(ctx context.Context, a *domain.Achievement) error {
	const q = `
		INSERT INTO achievement (user_id, type, ref, unlocked_at) VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, type, ref) DO NOTHING`
	if _, err := r.db.Exec(ctx, q, a.UserID, string(a.Type), a.Ref, a.UnlockedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

var _ domain.Repository = (*PgRepository)(nil)
