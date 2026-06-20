// Package infrastructure provides the Postgres adapter for the user repository.
package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	userdomain "github.com/son-ngo/edu-app/internal/user/domain"
)

const pgUniqueViolation = "23505"

// PgUserRepo implements userdomain.Repository over Postgres.
type PgUserRepo struct {
	db *pgxpool.Pool
}

// NewPgUserRepo builds the repository.
func NewPgUserRepo(db *pgxpool.Pool) *PgUserRepo { return &PgUserRepo{db: db} }

// Create inserts a profile, mapping a duplicate id/email to ErrConflict so the
// event handler can treat replays idempotently.
func (r *PgUserRepo) Create(ctx context.Context, u *userdomain.User) error {
	const q = `
		INSERT INTO users (id, email, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, q, u.ID, u.Email, u.DisplayName, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return shared.ErrConflict.WithMessage("user already exists")
		}
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

// FindByID returns a profile, or ErrNotFound.
func (r *PgUserRepo) FindByID(ctx context.Context, id string) (*userdomain.User, error) {
	const q = `SELECT id, email, display_name, created_at, updated_at FROM users WHERE id = $1`
	var u userdomain.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &u, nil
}

// Update persists profile changes.
func (r *PgUserRepo) Update(ctx context.Context, u *userdomain.User) error {
	const q = `UPDATE users SET display_name = $2, updated_at = $3 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, u.ID, u.DisplayName, u.UpdatedAt)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}
