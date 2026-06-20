package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/studyplan/domain"
)

// PgRepository implements domain.Repository over Postgres.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository builds the repository.
func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

// Save persists a plan, its milestones, and milestone topics atomically.
func (r *PgRepository) Save(ctx context.Context, p *domain.StudyPlan) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertPlan = `
		INSERT INTO study_plan (id, user_id, subject_id, level, start_date, target_date, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	if _, err := tx.Exec(ctx, insertPlan, p.ID, p.UserID, p.SubjectID, p.Level, p.StartDate, p.TargetDate, p.CreatedAt); err != nil {
		return shared.ErrInternal.WithCause(err)
	}

	for _, m := range p.Milestones {
		const insertM = `INSERT INTO milestone (id, plan_id, week_number, due_date) VALUES ($1,$2,$3,$4)`
		if _, err := tx.Exec(ctx, insertM, m.ID, p.ID, m.WeekNumber, m.DueDate); err != nil {
			return shared.ErrInternal.WithCause(err)
		}
		for i, topicID := range m.TopicIDs {
			const insertMT = `INSERT INTO milestone_topic (milestone_id, topic_id, order_index) VALUES ($1,$2,$3)`
			if _, err := tx.Exec(ctx, insertMT, m.ID, topicID, i); err != nil {
				return shared.ErrInternal.WithCause(err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// GetByID loads a plan with its milestones and topics.
func (r *PgRepository) GetByID(ctx context.Context, id string) (*domain.StudyPlan, error) {
	const q = `SELECT id, user_id, subject_id, level, start_date, target_date, created_at FROM study_plan WHERE id = $1`
	var p domain.StudyPlan
	err := r.db.QueryRow(ctx, q, id).Scan(&p.ID, &p.UserID, &p.SubjectID, &p.Level, &p.StartDate, &p.TargetDate, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("study plan not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	milestones, err := r.loadMilestones(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	p.Milestones = milestones[id]
	return &p, nil
}

// ListByUser loads a user's plans with milestones and topics.
func (r *PgRepository) ListByUser(ctx context.Context, userID string) ([]domain.StudyPlan, error) {
	const q = `SELECT id, user_id, subject_id, level, start_date, target_date, created_at FROM study_plan WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	var plans []domain.StudyPlan
	var ids []string
	for rows.Next() {
		var p domain.StudyPlan
		if err := rows.Scan(&p.ID, &p.UserID, &p.SubjectID, &p.Level, &p.StartDate, &p.TargetDate, &p.CreatedAt); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		plans = append(plans, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	if len(ids) == 0 {
		return plans, nil
	}

	milestones, err := r.loadMilestones(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range plans {
		plans[i].Milestones = milestones[plans[i].ID]
	}
	return plans, nil
}

// loadMilestones loads milestones (with topics) for many plans, keyed by plan id.
func (r *PgRepository) loadMilestones(ctx context.Context, planIDs []string) (map[string][]domain.Milestone, error) {
	const mq = `SELECT id, plan_id, week_number, due_date FROM milestone WHERE plan_id = ANY($1) ORDER BY week_number`
	rows, err := r.db.Query(ctx, mq, planIDs)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := map[string][]domain.Milestone{}
	idx := map[string]*domain.Milestone{} // milestone id -> pointer for topic attach
	order := map[string][]string{}        // plan id -> milestone ids in order
	for rows.Next() {
		var m domain.Milestone
		var planID string
		if err := rows.Scan(&m.ID, &planID, &m.WeekNumber, &m.DueDate); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out[planID] = append(out[planID], m)
		order[planID] = append(order[planID], m.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}

	// Build index into the slices for topic attachment.
	var milestoneIDs []string
	for planID := range out {
		for i := range out[planID] {
			idx[out[planID][i].ID] = &out[planID][i]
			milestoneIDs = append(milestoneIDs, out[planID][i].ID)
		}
	}
	if len(milestoneIDs) == 0 {
		return out, nil
	}

	const tq = `SELECT milestone_id, topic_id FROM milestone_topic WHERE milestone_id = ANY($1) ORDER BY order_index`
	trows, err := r.db.Query(ctx, tq, milestoneIDs)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer trows.Close()
	for trows.Next() {
		var milestoneID, topicID string
		if err := trows.Scan(&milestoneID, &topicID); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		if m := idx[milestoneID]; m != nil {
			m.TopicIDs = append(m.TopicIDs, topicID)
		}
	}
	return out, trows.Err()
}

var _ domain.Repository = (*PgRepository)(nil)
