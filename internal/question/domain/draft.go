package domain

import (
	"context"
	"time"
)

// DraftStatus is the review state of a parsed question draft.
type DraftStatus string

const (
	DraftPending   DraftStatus = "DRAFT"     // produced by the parse worker
	DraftPublished DraftStatus = "PUBLISHED" // reviewed and promoted to a Question
)

// QuestionDraftOption is a parsed answer option awaiting review.
type QuestionDraftOption struct {
	ID                string
	QuestionDraftID   string
	OptionLabel       string // e.g. "A", "B"
	OptionText        string
	IsCorrectInferred bool
	OrderIndex        int
}

// QuestionDraft is a parsed-but-unreviewed question from an uploaded PDF. It is
// written by the Python parse worker and promoted to a Question on admin review.
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

// DraftRepository persists question drafts (worker writes; admin reviews).
type DraftRepository interface {
	ListByAsset(ctx context.Context, assetID string) ([]QuestionDraft, error)
	GetByID(ctx context.Context, id string) (*QuestionDraft, error)
	UpdateDraft(ctx context.Context, id, stem, explanation string, at time.Time) error
	UpdateOption(ctx context.Context, optionID, text string, isCorrect bool) error
	MarkPublished(ctx context.Context, id, questionID, reviewedBy string, at time.Time) error
}
