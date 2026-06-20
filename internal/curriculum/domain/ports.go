package domain

import "context"

// Repository persists and retrieves the curriculum hierarchy.
type Repository interface {
	CreateSubject(ctx context.Context, s *Subject) error
	ListSubjects(ctx context.Context) ([]Subject, error)
	GetSubject(ctx context.Context, id string) (*Subject, error)

	CreateChapter(ctx context.Context, c *Chapter) error
	ListChaptersBySubject(ctx context.Context, subjectID string) ([]Chapter, error)
	GetChapter(ctx context.Context, id string) (*Chapter, error)

	CreateTopic(ctx context.Context, t *Topic) error
	ListTopicsByChapter(ctx context.Context, chapterID string) ([]Topic, error)
	// ListTopicsBySubject returns all topics under a subject (across chapters),
	// ordered by chapter then topic order. Used by placement/studyplan.
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]Topic, error)
	GetTopic(ctx context.Context, id string) (*Topic, error)
}
