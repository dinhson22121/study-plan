package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/curriculum/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository { return &PgRepository{db: db} }

func (r *PgRepository) CreateSubject(ctx context.Context, s *domain.Subject) error {
	const q = `INSERT INTO subject (id, code, name, grade_level) VALUES ($1, $2, $3, $4)`
	if _, err := r.db.Exec(ctx, q, s.ID, s.Code, s.Name, s.GradeLevel); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) ListSubjects(ctx context.Context) ([]domain.Subject, error) {
	const q = `SELECT id, code, name, grade_level FROM subject ORDER BY code`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.Subject
	for rows.Next() {
		var s domain.Subject
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.GradeLevel); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PgRepository) GetSubject(ctx context.Context, id string) (*domain.Subject, error) {
	const q = `SELECT id, code, name, grade_level FROM subject WHERE id = $1`
	var s domain.Subject
	err := r.db.QueryRow(ctx, q, id).Scan(&s.ID, &s.Code, &s.Name, &s.GradeLevel)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("subject not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &s, nil
}

func (r *PgRepository) CreateChapter(ctx context.Context, c *domain.Chapter) error {
	const q = `INSERT INTO chapter (id, subject_id, title, order_index) VALUES ($1, $2, $3, $4)`
	if _, err := r.db.Exec(ctx, q, c.ID, c.SubjectID, c.Title, c.OrderIndex); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) ListChaptersBySubject(ctx context.Context, subjectID string) ([]domain.Chapter, error) {
	const q = `SELECT id, subject_id, title, order_index FROM chapter WHERE subject_id = $1 ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, subjectID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.Chapter
	for rows.Next() {
		var c domain.Chapter
		if err := rows.Scan(&c.ID, &c.SubjectID, &c.Title, &c.OrderIndex); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PgRepository) GetChapter(ctx context.Context, id string) (*domain.Chapter, error) {
	const q = `SELECT id, subject_id, title, order_index FROM chapter WHERE id = $1`
	var c domain.Chapter
	err := r.db.QueryRow(ctx, q, id).Scan(&c.ID, &c.SubjectID, &c.Title, &c.OrderIndex)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("chapter not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &c, nil
}

func (r *PgRepository) CreateTopic(ctx context.Context, t *domain.Topic) error {
	const q = `INSERT INTO topic (id, chapter_id, title, order_index) VALUES ($1, $2, $3, $4)`
	if _, err := r.db.Exec(ctx, q, t.ID, t.ChapterID, t.Title, t.OrderIndex); err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgRepository) ListTopicsByChapter(ctx context.Context, chapterID string) ([]domain.Topic, error) {
	const q = `SELECT id, chapter_id, title, order_index FROM topic WHERE chapter_id = $1 ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, chapterID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.Topic
	for rows.Next() {
		var t domain.Topic
		if err := rows.Scan(&t.ID, &t.ChapterID, &t.Title, &t.OrderIndex); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *PgRepository) ListTopicsBySubject(ctx context.Context, subjectID string) ([]domain.Topic, error) {
	const q = `
		SELECT t.id, t.chapter_id, t.title, t.order_index
		FROM topic t
		JOIN chapter c ON c.id = t.chapter_id
		WHERE c.subject_id = $1
		ORDER BY c.order_index, t.order_index`
	rows, err := r.db.Query(ctx, q, subjectID)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.Topic
	for rows.Next() {
		var t domain.Topic
		if err := rows.Scan(&t.ID, &t.ChapterID, &t.Title, &t.OrderIndex); err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *PgRepository) GetTopic(ctx context.Context, id string) (*domain.Topic, error) {
	const q = `SELECT id, chapter_id, title, order_index FROM topic WHERE id = $1`
	var t domain.Topic
	err := r.db.QueryRow(ctx, q, id).Scan(&t.ID, &t.ChapterID, &t.Title, &t.OrderIndex)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("topic not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &t, nil
}

var _ domain.Repository = (*PgRepository)(nil)
