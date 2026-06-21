package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/son-ngo/edu-app/internal/content/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type AdminUploadService struct {
	assets      domain.AssetRepository
	jobs        domain.ParseJobRepository
	storage     domain.ObjectStorage
	maxFileSize int64
	now         func() time.Time
	newID       func() string
}

func NewAdminUploadService(assets domain.AssetRepository, jobs domain.ParseJobRepository, storage domain.ObjectStorage, maxFileSize int64) *AdminUploadService {
	return &AdminUploadService{
		assets: assets, jobs: jobs, storage: storage, maxFileSize: maxFileSize,
		now: time.Now, newID: uuid.NewString,
	}
}

type InitInput struct {
	UploadedBy     string
	Filename       string
	ContentType    string
	FileSize       int64
	ChecksumSHA256 string
}

type InitResult struct {
	Asset  *domain.UploadedAsset
	Upload domain.PresignedUpload
}

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

	asset, err := domain.NewPendingAsset(assetID, objectKey, s.storage.Bucket(), in.Filename, in.ContentType, in.UploadedBy, in.ChecksumSHA256, now)
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

type CompleteResult struct {
	Asset      *domain.UploadedAsset
	ParseJobID string
}

func (s *AdminUploadService) CompleteUpload(ctx context.Context, userID, assetID string) (*CompleteResult, error) {
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}
	if asset.Status == domain.AssetDeleted {
		return nil, shared.ErrConflict.WithMessage("asset has been deleted")
	}
	if asset.Status == domain.AssetUploaded || asset.Status == domain.AssetVerified {
		return s.idempotentComplete(ctx, asset)
	}

	found, size, contentType, err := s.storage.Head(ctx, asset.ObjectKey)
	if err != nil {
		return nil, shared.ErrInternal.WithCause(err)
	}
	if !found {
		return nil, shared.ErrValidation.WithMessage("object not found in storage; upload first")
	}
	if asset.FileSize > 0 && size != asset.FileSize {
		return nil, shared.ErrValidation.WithMessage("uploaded object size does not match the declared file size")
	}
	if contentType != "" && !sameContentType(asset.ContentType, contentType) {
		return nil, shared.ErrValidation.WithMessage("uploaded object content type does not match the declared content type")
	}
	if asset.ChecksumSHA256 != "" {
		body, err := s.storage.ReadAll(ctx, asset.ObjectKey)
		if err != nil {
			return nil, shared.ErrInternal.WithCause(err)
		}
		sum := sha256.Sum256(body)
		if hex.EncodeToString(sum[:]) != asset.ChecksumSHA256 {
			return nil, shared.ErrValidation.WithMessage("uploaded object checksum does not match checksum_sha256")
		}
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

func sameContentType(expected, actual string) bool {
	expected = mediaType(expected)
	actual = mediaType(actual)
	if expected == "" || actual == "" {
		return true
	}
	return expected == actual
}

func mediaType(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(v)
	if err == nil {
		return strings.ToLower(mt)
	}
	return strings.ToLower(v)
}

func (s *AdminUploadService) RetryParse(ctx context.Context, userID, assetID string) (*domain.ParseJob, error) {
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}
	if asset.Status != domain.AssetUploaded && asset.Status != domain.AssetVerified {
		return nil, shared.ErrConflict.WithMessage("asset must be uploaded before parsing")
	}
	job := domain.NewParseJob(s.newID(), asset.ID, userID, s.now())
	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *AdminUploadService) ListAssets(ctx context.Context, status string, limit, offset int) ([]domain.UploadedAsset, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.assets.List(ctx, domain.AssetFilter{Status: domain.AssetStatus(status), Limit: limit, Offset: offset})
}

func (s *AdminUploadService) GetAsset(ctx context.Context, id string) (*domain.UploadedAsset, error) {
	return s.assets.GetByID(ctx, id)
}

func (s *AdminUploadService) ListParseJobs(ctx context.Context, assetID string) ([]domain.ParseJob, error) {
	if _, err := s.assets.GetByID(ctx, assetID); err != nil {
		return nil, err
	}
	return s.jobs.ListByAsset(ctx, assetID)
}

func (s *AdminUploadService) LinkEntity(ctx context.Context, assetID, entityType, entityID string) error {
	if strings.TrimSpace(entityID) == "" {
		return shared.ErrValidation.WithMessage("entity_id is required")
	}
	typed := domain.AssetEntityType(strings.ToUpper(strings.TrimSpace(entityType)))
	if !typed.Valid() {
		return shared.ErrValidation.WithMessage("invalid entity_type")
	}
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return err
	}
	if asset.Status != domain.AssetUploaded && asset.Status != domain.AssetVerified {
		return shared.ErrConflict.WithMessage("asset must be uploaded before linking")
	}
	return s.assets.LinkEntity(ctx, assetID, typed, entityID)
}

func (s *AdminUploadService) DeleteAsset(ctx context.Context, id string) error {
	if _, err := s.assets.GetByID(ctx, id); err != nil {
		return err
	}
	return s.assets.SoftDelete(ctx, id, s.now())
}
