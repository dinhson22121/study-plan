package domain

import "context"

type Repository interface {
	CreateLesson(ctx context.Context, l *Lesson) error
	GetLesson(ctx context.Context, id string) (*Lesson, error)
	ListByTopic(ctx context.Context, topicID string) ([]Lesson, error)
}
