package domain

import "context"

// Repository persists and retrieves user goals.
type Repository interface {
	// Upsert creates or replaces the user's goal (one goal per user).
	Upsert(ctx context.Context, g *Goal) error
	// GetByUserID returns the user's goal, or ErrNotFound.
	GetByUserID(ctx context.Context, userID string) (*Goal, error)
}
