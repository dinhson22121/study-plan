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
	ID            string
	AssetID       string
	Status        ParseJobStatus
	ParserVersion string
	AttemptCount  int
	ErrorMessage  string
	ClaimedBy     string
	ClaimedAt     *time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	RawText       string
	CreatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewParseJob(id, assetID, createdBy string, now time.Time) *ParseJob {
	return &ParseJob{
		ID: id, AssetID: assetID, Status: ParseQueued,
		CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
	}
}
