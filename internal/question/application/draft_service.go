package application

import (
	"context"
	"time"

	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// DraftService implements review/publish of parser-produced question drafts.
type DraftService struct {
	drafts    domain.DraftRepository
	questions *Service
	now       func() time.Time
}

// NewDraftService builds the service.
func NewDraftService(drafts domain.DraftRepository, questions *Service) *DraftService {
	return &DraftService{drafts: drafts, questions: questions, now: time.Now}
}

// ListByAsset returns the drafts parsed from an uploaded asset.
func (s *DraftService) ListByAsset(ctx context.Context, assetID string) ([]domain.QuestionDraft, error) {
	return s.drafts.ListByAsset(ctx, assetID)
}

// Get returns one draft with its options.
func (s *DraftService) Get(ctx context.Context, id string) (*domain.QuestionDraft, error) {
	return s.drafts.GetByID(ctx, id)
}

// UpdateDraft edits a draft's stem/explanation during review.
func (s *DraftService) UpdateDraft(ctx context.Context, id, stem, explanation string) error {
	if stem == "" {
		return shared.ErrValidation.WithMessage("stem cannot be empty")
	}
	return s.drafts.UpdateDraft(ctx, id, stem, explanation, s.now())
}

// UpdateOption edits a draft option during review.
func (s *DraftService) UpdateOption(ctx context.Context, optionID, text string, isCorrect bool) error {
	if text == "" {
		return shared.ErrValidation.WithMessage("option text cannot be empty")
	}
	return s.drafts.UpdateOption(ctx, optionID, text, isCorrect)
}

// PublishInput is the publish-draft command.
type PublishInput struct {
	DraftID    string
	TopicID    string
	Difficulty string
	ReviewedBy string
}

// Publish promotes a reviewed draft into a real Question (via the question
// service, which enforces MCQ validity) and records the link on the draft.
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
