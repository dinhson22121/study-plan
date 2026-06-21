package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

const pgUniqueViolation = "23505"

type PgCredentialRepo struct {
	db *pgxpool.Pool
}

func NewPgCredentialRepo(db *pgxpool.Pool) *PgCredentialRepo {
	return &PgCredentialRepo{db: db}
}

func (r *PgCredentialRepo) Create(ctx context.Context, c *authdomain.UserCredential) error {
	const q = `
		INSERT INTO user_credential (user_id, email, password_hash, role)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, q, c.UserID, c.Email, c.PasswordHash, string(c.Role))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrConflict.WithMessage("email already registered")
		}
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgCredentialRepo) FindByEmail(ctx context.Context, email string) (*authdomain.UserCredential, error) {
	const q = `SELECT user_id, email, password_hash, role FROM user_credential WHERE email = $1`
	return r.scanOne(ctx, q, email)
}

func (r *PgCredentialRepo) FindByUserID(ctx context.Context, userID string) (*authdomain.UserCredential, error) {
	const q = `SELECT user_id, email, password_hash, role FROM user_credential WHERE user_id = $1`
	return r.scanOne(ctx, q, userID)
}

func (r *PgCredentialRepo) scanOne(ctx context.Context, q string, arg any) (*authdomain.UserCredential, error) {
	var c authdomain.UserCredential
	var role string
	err := r.db.QueryRow(ctx, q, arg).Scan(&c.UserID, &c.Email, &c.PasswordHash, &role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.ErrInternal.WithCause(err)
	}
	c.Role = authdomain.Role(role)
	return &c, nil
}
