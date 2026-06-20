package domain

import (
	"context"
	"time"
)

// AssetFilter narrows an asset listing.
type AssetFilter struct {
	Status AssetStatus // optional
	Limit  int
	Offset int
}

// AssetRepository persists uploaded-asset metadata.
type AssetRepository interface {
	Create(ctx context.Context, a *UploadedAsset) error
	GetByID(ctx context.Context, id string) (*UploadedAsset, error)
	List(ctx context.Context, f AssetFilter) ([]UploadedAsset, int, error)
	MarkUploaded(ctx context.Context, id string, size int64, verifiedAt time.Time) error
	SoftDelete(ctx context.Context, id string, at time.Time) error
}

// ParseJobRepository persists parse jobs.
type ParseJobRepository interface {
	Create(ctx context.Context, j *ParseJob) error
	ListByAsset(ctx context.Context, assetID string) ([]ParseJob, error)
}

// PresignedUpload is the storage-agnostic presign result returned to clients.
type PresignedUpload struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

// ObjectStorage abstracts the S3-compatible store (keeps domain free of the SDK).
type ObjectStorage interface {
	Bucket() string
	PresignPut(ctx context.Context, key, contentType string) (PresignedUpload, error)
	Head(ctx context.Context, key string) (found bool, size int64, contentType string, err error)
}
