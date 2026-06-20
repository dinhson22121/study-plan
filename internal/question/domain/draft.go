package domain

import (
	"context"
	"time"
)

type DraftStatus string

const (
	DraftPending   DraftStatus = "DRAFT"
	DraftPublished DraftStatus = "PUBLISHED"
)

type QuestionDraftOption struct {
	ID                string
	QuestionDraftID   string
	OptionLabel       string
	OptionText        string
	IsCorrectInferred bool
	OrderIndex        int
}

type QuestionDraft struct {
	ID                  string
	AssetID             string
	ParseJobID          string
	QuestionNumber      int
	QuestionType        string
	Stem                string
	ExplanationRaw      string
	AnswerKeyRaw        string
	ParseConfidence     float64
	Status              DraftStatus
	ReviewedBy          string
	ReviewedAt          *time.Time
	PublishedQuestionID string
	Options             []QuestionDraftOption
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type DraftRepository interface {
	ListByAsset(ctx context.Context, assetID string) ([]QuestionDraft, error)
	GetByID(ctx context.Context, id string) (*QuestionDraft, error)
	UpdateDraft(ctx context.Context, id, stem, explanation string, at time.Time) error
	UpdateOption(ctx context.Context, optionID, text string, isCorrect bool) error
	MarkPublished(ctx context.Context, id, questionID, reviewedBy string, at time.Time) error
}
