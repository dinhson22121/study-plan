package infrastructure

import (
	"context"

	curriculumapp "github.com/son-ngo/edu-app/internal/curriculum/application"
)

type TopicTitleAdapter struct{ curriculum *curriculumapp.Service }

func NewTopicTitleAdapter(c *curriculumapp.Service) *TopicTitleAdapter {
	return &TopicTitleAdapter{curriculum: c}
}

func (a *TopicTitleAdapter) Title(ctx context.Context, topicID string) (string, error) {
	t, err := a.curriculum.GetTopic(ctx, topicID)
	if err != nil {
		return "", err
	}
	return t.Title, nil
}
