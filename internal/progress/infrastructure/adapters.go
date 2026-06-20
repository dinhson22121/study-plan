// Package infrastructure provides the progress adapters (topic titles via
// curriculum) and the Postgres repository.
package infrastructure

import (
	"context"

	curriculumapp "github.com/son-ngo/edu-app/internal/curriculum/application"
)

// TopicTitleAdapter implements progress's TopicTitleSource via curriculum.
type TopicTitleAdapter struct{ curriculum *curriculumapp.Service }

// NewTopicTitleAdapter builds the adapter.
func NewTopicTitleAdapter(c *curriculumapp.Service) *TopicTitleAdapter {
	return &TopicTitleAdapter{curriculum: c}
}

// Title returns the topic's title (used in the achievement push copy).
func (a *TopicTitleAdapter) Title(ctx context.Context, topicID string) (string, error) {
	t, err := a.curriculum.GetTopic(ctx, topicID)
	if err != nil {
		return "", err
	}
	return t.Title, nil
}
