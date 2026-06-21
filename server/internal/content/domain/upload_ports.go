package domain

import (
	"context"
	"time"
)

type AssetFilter struct {
	Status AssetStatus
	Limit  int
	Offset int
}

type AssetRepository interface {
	Create(ctx context.Context, a *UploadedAsset) error
	GetByID(ctx context.Context, id string) (*UploadedAsset, error)
	List(ctx context.Context, f AssetFilter) ([]UploadedAsset, int, error)
	MarkUploaded(ctx context.Context, id string, size int64, verifiedAt time.Time) error
	LinkEntity(ctx context.Context, id string, entityType AssetEntityType, entityID string) error
	SoftDelete(ctx context.Context, id string, at time.Time) error
}

type ParseJobRepository interface {
	Create(ctx context.Context, j *ParseJob) error
	ListByAsset(ctx context.Context, assetID string) ([]ParseJob, error)
}

type PresignedUpload struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

type ObjectStorage interface {
	Bucket() string
	PresignPut(ctx context.Context, key, contentType string) (PresignedUpload, error)
	Head(ctx context.Context, key string) (found bool, size int64, contentType string, err error)
	ReadAll(ctx context.Context, key string) ([]byte, error)
}
