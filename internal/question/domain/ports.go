package domain

import "context"

// ListFilter narrows a question query.
type ListFilter struct {
	TopicID    string
	Difficulty Difficulty // optional; empty means any
	Limit      int
}

// Repository persists and retrieves questions with their options.
type Repository interface {
	Create(ctx context.Context, q *Question) error
	GetByID(ctx context.Context, id string) (*Question, error)
	List(ctx context.Context, f ListFilter) ([]Question, error)
}
