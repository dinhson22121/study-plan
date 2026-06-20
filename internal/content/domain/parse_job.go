package domain

import "time"

// ParseJobStatus is the lifecycle state of a PDF parse job (driven by the Python
// worker; Go only creates jobs and reads their status).
type ParseJobStatus string

const (
	ParseQueued         ParseJobStatus = "QUEUED"
	ParseProcessing     ParseJobStatus = "PROCESSING"
	ParseParsed         ParseJobStatus = "PARSED"
	ParseReviewRequired ParseJobStatus = "REVIEW_REQUIRED"
	ParseFailed         ParseJobStatus = "FAILED"
)

// ParseJob represents one attempt to parse an uploaded PDF into draft questions.
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

// NewParseJob builds a QUEUED job for an asset.
func NewParseJob(id, assetID, createdBy string, now time.Time) *ParseJob {
	return &ParseJob{
		ID: id, AssetID: assetID, Status: ParseQueued,
		CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
	}
}
