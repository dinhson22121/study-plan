package domain

import "time"

type ParseJobStatus string

const (
	ParseQueued         ParseJobStatus = "QUEUED"
	ParseProcessing     ParseJobStatus = "PROCESSING"
	ParseParsed         ParseJobStatus = "PARSED"
	ParseReviewRequired ParseJobStatus = "REVIEW_REQUIRED"
	ParseFailed         ParseJobStatus = "FAILED"
)

type ParseJob struct {
	ID            string         `json:"id"`
	AssetID       string         `json:"asset_id"`
	Status        ParseJobStatus `json:"status"`
	ParserVersion string         `json:"parser_version"`
	AttemptCount  int            `json:"attempt_count"`
	ErrorMessage  string         `json:"error_message"`
	ClaimedBy     string         `json:"claimed_by"`
	ClaimedAt     *time.Time     `json:"claimed_at"`
	StartedAt     *time.Time     `json:"started_at"`
	FinishedAt    *time.Time     `json:"finished_at"`
	RawText       string         `json:"-"`
	CreatedBy     string         `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func NewParseJob(id, assetID, createdBy string, now time.Time) *ParseJob {
	return &ParseJob{
		ID: id, AssetID: assetID, Status: ParseQueued,
		CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
	}
}
