package domain

import "context"

// Repository persists and retrieves lessons with their content items.
type Repository interface {
	CreateLesson(ctx context.Context, l *Lesson) error
	GetLesson(ctx context.Context, id string) (*Lesson, error)
	ListByTopic(ctx context.Context, topicID string) ([]Lesson, error)
}
