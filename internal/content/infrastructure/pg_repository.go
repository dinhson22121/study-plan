// Package infrastructure provides the Postgres adapter for the content
// repository.
package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/content/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const pgForeignKeyViolation = "23503"

// PgRepository implements domain.Repository over Postgres.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository builds the repository.
func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

// CreateLesson inserts a lesson and its items atomically. A reference to a
// missing topic surfaces as a validation error.
func (r *PgRepository) CreateLesson(ctx context.Context, l *domain.Lesson) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertL = `INSERT INTO lesson (id, topic_id, title, order_index) VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(ctx, insertL, l.ID, l.TopicID, l.Title, l.OrderIndex); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return shared.ErrValidation.WithMessage("topic does not exist")
		}
		return shared.ErrInternal.WithCause(err)
	}

	const insertI = `INSERT INTO content_item (id, lesson_id, kind, url, body, order_index) VALUES ($1, $2, $3, $4, $5, $6)`
	for _, it := range l.Items {
		if _, err := tx.Exec(ctx, insertI, it.ID, l.ID, string(it.Kind), it.URL, it.Body, it.OrderIndex); err != nil {
			return shared.ErrInternal.WithCause(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// GetLesson returns a lesson with its items ordered.
func (r *PgRepository) GetLesson(ctx context.Context, id string) (*domain.Lesson, error) {
	const q = `SELECT id, topic_id, title, order_index FROM lesson WHERE id = $1`
	var l domain.Lesson
	err := r.db.QueryRow(ctx, q, id).Scan(&l.ID, &l.TopicID, &l.Title, &l.OrderIndex)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("lesson not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	items, err := r.loadItems(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	l.Items = items[id]
	return &l, nil
}

// ListByTopic returns lessons of a topic with their items.
func (r *PgRepository) ListByTopic(ctx context.Context, topicID string) ([]domain.Lesson, error) {
	const q = `SELECT id, topic_id, title, order_index FROM lesson WHERE topic_id = $1 ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, topicID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	var lessons []domain.Lesson
	var ids []string
	for rows.Next() {
		var l domain.Lesson
		if err := rows.Scan(&l.ID, &l.TopicID, &l.Title, &l.OrderIndex); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		lessons = append(lessons, l)
		ids = append(ids, l.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	if len(ids) == 0 {
		return lessons, nil
	}

	items, err := r.loadItems(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range lessons {
		lessons[i].Items = items[lessons[i].ID]
	}
	return lessons, nil
}

func (r *PgRepository) loadItems(ctx context.Context, lessonIDs []string) (map[string][]domain.ContentItem, error) {
	const q = `
		SELECT id, lesson_id, kind, url, body, order_index
		FROM content_item WHERE lesson_id = ANY($1) ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, lessonIDs)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := map[string][]domain.ContentItem{}
	for rows.Next() {
		var it domain.ContentItem
		var lessonID, kind string
		if err := rows.Scan(&it.ID, &lessonID, &kind, &it.URL, &it.Body, &it.OrderIndex); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		it.Kind = domain.ContentKind(kind)
		out[lessonID] = append(out[lessonID], it)
	}
	return out, rows.Err()
}

var _ domain.Repository = (*PgRepository)(nil)
