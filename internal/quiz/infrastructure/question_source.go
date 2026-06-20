package infrastructure

import (
	"context"

	questionapp "github.com/son-ngo/edu-app/internal/question/application"
	"github.com/son-ngo/edu-app/internal/quiz/domain"
)

type QuestionSourceAdapter struct {
	questions *questionapp.Service
}

func NewQuestionSourceAdapter(questions *questionapp.Service) *QuestionSourceAdapter {
	return &QuestionSourceAdapter{questions: questions}
}

func (a *QuestionSourceAdapter) SampleForTopic(ctx context.Context, topicID string, limit int) ([]string, error) {
	qs, err := a.questions.List(ctx, topicID, "", limit)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(qs))
	for _, q := range qs {
		ids = append(ids, q.ID)
	}
	return ids, nil
}

func (a *QuestionSourceAdapter) Details(ctx context.Context, questionIDs []string) (map[string]domain.QuestionDetail, error) {
	out := make(map[string]domain.QuestionDetail, len(questionIDs))
	for _, id := range questionIDs {
		q, err := a.questions.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		var correct []string
		for _, o := range q.Options {
			if o.IsCorrect {
				correct = append(correct, o.ID)
			}
		}
		out[id] = domain.QuestionDetail{CorrectOptionIDs: correct, Explanation: q.Explanation}
	}
	return out, nil
}
