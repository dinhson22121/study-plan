package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/content/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// AdminUploadService implements the admin upload + parse-job use cases.
type AdminUploadService struct {
	assets      domain.AssetRepository
	jobs        domain.ParseJobRepository
	storage     domain.ObjectStorage
	maxFileSize int64
	now         func() time.Time
	newID       func() string
}

// NewAdminUploadService builds the service.
func NewAdminUploadService(assets domain.AssetRepository, jobs domain.ParseJobRepository, storage domain.ObjectStorage, maxFileSize int64) *AdminUploadService {
	return &AdminUploadService{
		assets: assets, jobs: jobs, storage: storage, maxFileSize: maxFileSize,
		now: time.Now, newID: uuid.NewString,
	}
}

// InitInput is the init-upload command.
type InitInput struct {
	UploadedBy  string
	Filename    string
	ContentType string
	FileSize    int64
}

// InitResult carries the created asset and the presigned upload instructions.
type InitResult struct {
	Asset  *domain.UploadedAsset
	Upload domain.PresignedUpload
}

// InitUpload validates the request, creates a PENDING asset, and returns a
// presigned PUT URL for the admin client to upload directly to storage.
func (s *AdminUploadService) InitUpload(ctx context.Context, in InitInput) (*InitResult, error) {
	if in.FileSize <= 0 {
		return nil, shared.ErrValidation.WithMessage("file size must be positive")
	}
	if in.FileSize > s.maxFileSize {
		return nil, shared.ErrValidation.WithMessage("file exceeds the maximum allowed size")
	}

	now := s.now()
	assetID := s.newID()
	objectKey := domain.BuildObjectKey(assetID, in.Filename, now)

	asset, err := domain.NewPendingAsset(assetID, objectKey, s.storage.Bucket(), in.Filename, in.ContentType, in.UploadedBy, now)
	if err != nil {
		return nil, err
	}
	asset.FileSize = in.FileSize
	if err := s.assets.Create(ctx, asset); err != nil {
		return nil, err
	}

	presign, err := s.storage.PresignPut(ctx, objectKey, in.ContentType)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	return &InitResult{Asset: asset, Upload: presign}, nil
}

// CompleteResult is returned after a successful upload completion.
type CompleteResult struct {
	Asset      *domain.UploadedAsset
	ParseJobID string
}

// CompleteUpload verifies the object exists in storage, marks the asset UPLOADED,
// and queues a parse job. It is idempotent: completing an already-UPLOADED asset
// returns its latest parse job without creating a duplicate.
func (s *AdminUploadService) CompleteUpload(ctx context.Context, userID, assetID string) (*CompleteResult, error) {
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}
	if asset.Status == domain.AssetDeleted {
		return nil, shared.ErrConflict.WithMessage("asset has been deleted")
	}
	if asset.Status == domain.AssetUploaded {
		return s.idempotentComplete(ctx, asset)
	}

	found, size, _, err := s.storage.Head(ctx, asset.ObjectKey)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	if !found {
		return nil, shared.ErrValidation.WithMessage("object not found in storage; upload first")
	}

	now := s.now()
	if err := s.assets.MarkUploaded(ctx, asset.ID, size, now); err != nil {
		return nil, err
	}
	asset.Status = domain.AssetUploaded
	asset.FileSize = size
	asset.VerifiedAt = &now

	job := domain.NewParseJob(s.newID(), asset.ID, userID, now)
	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, err
	}
	return &CompleteResult{Asset: asset, ParseJobID: job.ID}, nil
}

func (s *AdminUploadService) idempotentComplete(ctx context.Context, asset *domain.UploadedAsset) (*CompleteResult, error) {
	jobs, err := s.jobs.ListByAsset(ctx, asset.ID)
	if err != nil {
		return nil, err
	}
	res := &CompleteResult{Asset: asset}
	if len(jobs) > 0 {
		res.ParseJobID = jobs[0].ID
	}
	return res, nil
}

// RetryParse queues a new parse job for an already-uploaded asset.
func (s *AdminUploadService) RetryParse(ctx context.Context, userID, assetID string) (*domain.ParseJob, error) {
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}
	if asset.Status != domain.AssetUploaded {
		return nil, shared.ErrConflict.WithMessage("asset must be uploaded before parsing")
	}
	job := domain.NewParseJob(s.newID(), asset.ID, userID, s.now())
	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

// ListAssets returns a page of assets.
func (s *AdminUploadService) ListAssets(ctx context.Context, status string, limit, offset int) ([]domain.UploadedAsset, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.assets.List(ctx, domain.AssetFilter{Status: domain.AssetStatus(status), Limit: limit, Offset: offset})
}

// GetAsset returns one asset by id.
func (s *AdminUploadService) GetAsset(ctx context.Context, id string) (*domain.UploadedAsset, error) {
	return s.assets.GetByID(ctx, id)
}

// ListParseJobs returns an asset's parse-job history.
func (s *AdminUploadService) ListParseJobs(ctx context.Context, assetID string) ([]domain.ParseJob, error) {
	if _, err := s.assets.GetByID(ctx, assetID); err != nil {
		return nil, err
	}
	return s.jobs.ListByAsset(ctx, assetID)
}

// DeleteAsset soft-deletes an asset.
func (s *AdminUploadService) DeleteAsset(ctx context.Context, id string) error {
	if _, err := s.assets.GetByID(ctx, id); err != nil {
		return err
	}
	return s.assets.SoftDelete(ctx, id, s.now())
}
