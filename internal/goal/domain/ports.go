package domain

import "context"

type Repository interface {
	Upsert(ctx context.Context, g *Goal) error

	GetByUserID(ctx context.Context, userID string) (*Goal, error)
}
