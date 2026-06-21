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
	ID                string `json:"id"`
	QuestionDraftID   string `json:"question_draft_id"`
	OptionLabel       string `json:"option_label"`
	OptionText        string `json:"option_text"`
	IsCorrectInferred bool   `json:"is_correct_inferred"`
	OrderIndex        int    `json:"order_index"`
}

type QuestionDraft struct {
	ID                  string                `json:"id"`
	AssetID             string                `json:"asset_id"`
	ParseJobID          string                `json:"parse_job_id"`
	QuestionNumber      int                   `json:"question_number"`
	QuestionType        string                `json:"question_type"`
	Stem                string                `json:"stem"`
	ExplanationRaw      string                `json:"explanation_raw"`
	AnswerKeyRaw        string                `json:"answer_key_raw"`
	ParseConfidence     float64               `json:"parse_confidence"`
	Status              DraftStatus           `json:"status"`
	ReviewedBy          string                `json:"reviewed_by"`
	ReviewedAt          *time.Time            `json:"reviewed_at"`
	PublishedQuestionID string                `json:"published_question_id"`
	Options             []QuestionDraftOption `json:"options"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

type DraftRepository interface {
	ListByAsset(ctx context.Context, assetID string) ([]QuestionDraft, error)
	GetByID(ctx context.Context, id string) (*QuestionDraft, error)
	UpdateDraft(ctx context.Context, id, stem, explanation string, at time.Time) error
	UpdateOption(ctx context.Context, optionID, text string, isCorrect bool) error
	MarkPublished(ctx context.Context, id, questionID, reviewedBy string, at time.Time) error
}
