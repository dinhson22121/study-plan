package domain

import "context"

type ListFilter struct {
	TopicID    string
	Difficulty Difficulty
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, q *Question) error
	GetByID(ctx context.Context, id string) (*Question, error)
	List(ctx context.Context, f ListFilter) ([]Question, error)
}
