package application

import (
	"context"
	"time"

	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type DraftService struct {
	drafts    domain.DraftRepository
	questions *Service
	now       func() time.Time
}

func NewDraftService(drafts domain.DraftRepository, questions *Service) *DraftService {
	return &DraftService{drafts: drafts, questions: questions, now: time.Now}
}

func (s *DraftService) ListByAsset(ctx context.Context, assetID string) ([]domain.QuestionDraft, error) {
	return s.drafts.ListByAsset(ctx, assetID)
}

func (s *DraftService) Get(ctx context.Context, id string) (*domain.QuestionDraft, error) {
	return s.drafts.GetByID(ctx, id)
}

func (s *DraftService) UpdateDraft(ctx context.Context, id, stem, explanation string) error {
	if stem == "" {
		return shared.ErrValidation.WithMessage("stem cannot be empty")
	}
	return s.drafts.UpdateDraft(ctx, id, stem, explanation, s.now())
}

func (s *DraftService) UpdateOption(ctx context.Context, optionID, text string, isCorrect bool) error {
	if text == "" {
		return shared.ErrValidation.WithMessage("option text cannot be empty")
	}
	return s.drafts.UpdateOption(ctx, optionID, text, isCorrect)
}

type PublishInput struct {
	DraftID    string
	TopicID    string
	Difficulty string
	ReviewedBy string
}

type PublishByAssetInput struct {
	AssetID    string
	TopicID    string
	Difficulty string
	ReviewedBy string
}

func (s *DraftService) Publish(ctx context.Context, in PublishInput) (*domain.Question, error) {
	draft, err := s.drafts.GetByID(ctx, in.DraftID)
	if err != nil {
		return nil, err
	}
	if draft.Status == domain.DraftPublished {
		return nil, shared.ErrConflict.WithMessage("draft already published")
	}

	opts := make([]OptionInput, 0, len(draft.Options))
	for _, o := range draft.Options {
		opts = append(opts, OptionInput{Text: o.OptionText, IsCorrect: o.IsCorrectInferred})
	}
	question, err := s.questions.Create(ctx, CreateInput{
		TopicID:     in.TopicID,
		Type:        draft.QuestionType,
		Stem:        draft.Stem,
		Difficulty:  in.Difficulty,
		Explanation: draft.ExplanationRaw,
		Options:     opts,
	})
	if err != nil {
		return nil, err
	}

	if err := s.drafts.MarkPublished(ctx, draft.ID, question.ID, in.ReviewedBy, s.now()); err != nil {
		return nil, err
	}
	return question, nil
}

func (s *DraftService) PublishByAsset(ctx context.Context, in PublishByAssetInput) ([]*domain.Question, error) {
	drafts, err := s.drafts.ListByAsset(ctx, in.AssetID)
	if err != nil {
		return nil, err
	}
	if len(drafts) == 0 {
		return nil, shared.ErrNotFound.WithMessage("no drafts found for asset")
	}

	out := make([]*domain.Question, 0, len(drafts))
	for _, draft := range drafts {
		if draft.Status == domain.DraftPublished {
			continue
		}
		question, err := s.Publish(ctx, PublishInput{
			DraftID: draft.ID, TopicID: in.TopicID, Difficulty: in.Difficulty, ReviewedBy: in.ReviewedBy,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, question)
	}
	if len(out) == 0 {
		return nil, shared.ErrConflict.WithMessage("all drafts for asset have already been published")
	}
	return out, nil
}
