package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/question/domain"
)

type Service struct {
	repo  domain.Repository
	newID func() string
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo, newID: uuid.NewString}
}

type OptionInput struct {
	Text      string
	IsCorrect bool
}

type CreateInput struct {
	TopicID     string
	Type        string
	Stem        string
	Difficulty  string
	Explanation string
	Options     []OptionInput
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Question, error) {
	difficulty, err := domain.ParseDifficulty(in.Difficulty)
	if err != nil {
		return nil, err
	}

	options := make([]domain.AnswerOption, 0, len(in.Options))
	for i, o := range in.Options {
		options = append(options, domain.AnswerOption{
			ID:         s.newID(),
			Text:       o.Text,
			IsCorrect:  o.IsCorrect,
			OrderIndex: i,
		})
	}

	q, err := domain.NewQuestion(s.newID(), in.TopicID, domain.QuestionType(in.Type), in.Stem, difficulty, in.Explanation, options)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Question, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, topicID, difficulty string, limit int) ([]domain.Question, error) {
	f := domain.ListFilter{TopicID: topicID, Limit: limit}
	if difficulty != "" {
		d, err := domain.ParseDifficulty(difficulty)
		if err != nil {
			return nil, err
		}
		f.Difficulty = d
	}
	return s.repo.List(ctx, f)
}
