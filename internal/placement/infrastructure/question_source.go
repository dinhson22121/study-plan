// Package infrastructure provides the placement adapters: the question-bank
// source (bridging curriculum + question) and the Postgres repository.
package infrastructure

import (
	"context"

	curriculumapp "github.com/son-ngo/edu-app/internal/curriculum/application"
	questionapp "github.com/son-ngo/edu-app/internal/question/application"
)

// QuestionSourceAdapter implements placement's QuestionSource port by composing
// the curriculum and question services. This is the single integration point
// where placement reads from those contexts; placement's core logic depends only
// on the port, not these packages.
type QuestionSourceAdapter struct {
	curriculum *curriculumapp.Service
	questions  *questionapp.Service
}

// NewQuestionSourceAdapter builds the adapter.
func NewQuestionSourceAdapter(curriculum *curriculumapp.Service, questions *questionapp.Service) *QuestionSourceAdapter {
	return &QuestionSourceAdapter{curriculum: curriculum, questions: questions}
}

// SampleForSubject collects up to limit question ids spanning the subject's
// topics (in curriculum order).
func (a *QuestionSourceAdapter) SampleForSubject(ctx context.Context, subjectID string, limit int) ([]string, error) {
	topics, err := a.curriculum.ListTopicsBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, t := range topics {
		if len(ids) >= limit {
			break
		}
		qs, err := a.questions.List(ctx, t.ID, "", limit-len(ids))
		if err != nil {
			return nil, err
		}
		for _, q := range qs {
			ids = append(ids, q.ID)
			if len(ids) >= limit {
				break
			}
		}
	}
	return ids, nil
}

// CorrectOptions returns, per question id, the set of correct option ids.
func (a *QuestionSourceAdapter) CorrectOptions(ctx context.Context, questionIDs []string) (map[string]map[string]bool, error) {
	out := make(map[string]map[string]bool, len(questionIDs))
	for _, id := range questionIDs {
		q, err := a.questions.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		set := map[string]bool{}
		for _, o := range q.Options {
			if o.IsCorrect {
				set[o.ID] = true
			}
		}
		out[id] = set
	}
	return out, nil
}
