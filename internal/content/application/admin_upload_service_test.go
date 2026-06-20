package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/son-ngo/edu-app/internal/content/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

type fakeAssetRepo struct {
	assets map[string]*domain.UploadedAsset
}

func newFakeAssetRepo() *fakeAssetRepo {
	return &fakeAssetRepo{assets: map[string]*domain.UploadedAsset{}}
}

func (r *fakeAssetRepo) Create(_ context.Context, a *domain.UploadedAsset) error {
	cp := *a
	r.assets[a.ID] = &cp
	return nil
}
func (r *fakeAssetRepo) GetByID(_ context.Context, id string) (*domain.UploadedAsset, error) {
	if a, ok := r.assets[id]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, shared.ErrNotFound
}
func (r *fakeAssetRepo) List(_ context.Context, _ domain.AssetFilter) ([]domain.UploadedAsset, int, error) {
	var out []domain.UploadedAsset
	for _, a := range r.assets {
		out = append(out, *a)
	}
	return out, len(out), nil
}
func (r *fakeAssetRepo) MarkUploaded(_ context.Context, id string, size int64, at time.Time) error {
	a, ok := r.assets[id]
	if !ok {
		return shared.ErrNotFound
	}
	a.Status = domain.AssetUploaded
	a.FileSize = size
	a.VerifiedAt = &at
	return nil
}
func (r *fakeAssetRepo) SoftDelete(_ context.Context, id string, at time.Time) error {
	a, ok := r.assets[id]
	if !ok {
		return shared.ErrNotFound
	}
	a.Status = domain.AssetDeleted
	a.DeletedAt = &at
	return nil
}

type fakeJobRepo struct{ jobs []*domain.ParseJob }

func (r *fakeJobRepo) Create(_ context.Context, j *domain.ParseJob) error {
	r.jobs = append(r.jobs, j)
	return nil
}
func (r *fakeJobRepo) ListByAsset(_ context.Context, assetID string) ([]domain.ParseJob, error) {
	var out []domain.ParseJob
	for i := len(r.jobs) - 1; i >= 0; i-- {
		if r.jobs[i].AssetID == assetID {
			out = append(out, *r.jobs[i])
		}
	}
	return out, nil
}

type fakeStorage struct {
	found bool
	size  int64
}

func (s *fakeStorage) Bucket() string { return "test-bucket" }
func (s *fakeStorage) PresignPut(_ context.Context, key, _ string) (domain.PresignedUpload, error) {
	return domain.PresignedUpload{URL: "http://minio/" + key, Method: "PUT"}, nil
}
func (s *fakeStorage) Head(_ context.Context, _ string) (bool, int64, string, error) {
	return s.found, s.size, "application/pdf", nil
}

func newService(assets *fakeAssetRepo, jobs *fakeJobRepo, storage *fakeStorage) *AdminUploadService {
	svc := NewAdminUploadService(assets, jobs, storage, 20*1024*1024)
	svc.now = func() time.Time { return time.Unix(1000, 0).UTC() }
	seq := 0
	svc.newID = func() string { seq++; return "id" + string(rune('0'+seq)) }
	return svc
}

func TestInitUpload_CreatesPendingAssetAndPresign(t *testing.T) {
	assets, jobs, storage := newFakeAssetRepo(), &fakeJobRepo{}, &fakeStorage{}
	svc := newService(assets, jobs, storage)

	res, err := svc.InitUpload(context.Background(), InitInput{
		UploadedBy: "admin", Filename: "de.pdf", ContentType: "application/pdf", FileSize: 1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Asset.Status != domain.AssetPending || res.Upload.URL == "" {
		t.Fatalf("bad init result: %+v", res)
	}
	if assets.assets[res.Asset.ID] == nil {
		t.Fatalf("asset not stored")
	}
}

func TestInitUpload_RejectsOversizeAndNonPDF(t *testing.T) {
	svc := newService(newFakeAssetRepo(), &fakeJobRepo{}, &fakeStorage{})
	if _, err := svc.InitUpload(context.Background(), InitInput{UploadedBy: "admin", Filename: "x.pdf", ContentType: "application/pdf", FileSize: 999999999999}); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected oversize validation error")
	}
	if _, err := svc.InitUpload(context.Background(), InitInput{UploadedBy: "admin", Filename: "x.png", ContentType: "image/png", FileSize: 100}); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected non-pdf validation error")
	}
}

func TestCompleteUpload_VerifiesAndQueuesJob(t *testing.T) {
	assets, jobs, storage := newFakeAssetRepo(), &fakeJobRepo{}, &fakeStorage{found: true, size: 2048}
	svc := newService(assets, jobs, storage)
	initRes, _ := svc.InitUpload(context.Background(), InitInput{UploadedBy: "admin", Filename: "de.pdf", ContentType: "application/pdf", FileSize: 2048})

	res, err := svc.CompleteUpload(context.Background(), "admin", initRes.Asset.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ParseJobID == "" || res.Asset.Status != domain.AssetUploaded {
		t.Fatalf("complete result wrong: %+v", res)
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected 1 parse job, got %d", len(jobs.jobs))
	}
}

func TestCompleteUpload_ObjectMissing(t *testing.T) {
	assets, jobs, storage := newFakeAssetRepo(), &fakeJobRepo{}, &fakeStorage{found: false}
	svc := newService(assets, jobs, storage)
	initRes, _ := svc.InitUpload(context.Background(), InitInput{UploadedBy: "admin", Filename: "de.pdf", ContentType: "application/pdf", FileSize: 100})

	if _, err := svc.CompleteUpload(context.Background(), "admin", initRes.Asset.ID); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected validation error when object missing, got %v", err)
	}
}

func TestCompleteUpload_Idempotent(t *testing.T) {
	assets, jobs, storage := newFakeAssetRepo(), &fakeJobRepo{}, &fakeStorage{found: true, size: 100}
	svc := newService(assets, jobs, storage)
	initRes, _ := svc.InitUpload(context.Background(), InitInput{UploadedBy: "admin", Filename: "de.pdf", ContentType: "application/pdf", FileSize: 100})

	first, _ := svc.CompleteUpload(context.Background(), "admin", initRes.Asset.ID)
	second, err := svc.CompleteUpload(context.Background(), "admin", initRes.Asset.ID)
	if err != nil {
		t.Fatalf("idempotent complete should not error: %v", err)
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected no duplicate parse job, got %d", len(jobs.jobs))
	}
	if second.ParseJobID != first.ParseJobID {
		t.Fatalf("idempotent complete should return same job id")
	}
}

func TestRetryParse_RequiresUploaded(t *testing.T) {
	assets, jobs, storage := newFakeAssetRepo(), &fakeJobRepo{}, &fakeStorage{}
	svc := newService(assets, jobs, storage)
	initRes, _ := svc.InitUpload(context.Background(), InitInput{UploadedBy: "admin", Filename: "de.pdf", ContentType: "application/pdf", FileSize: 100})

	// Still PENDING -> retry should be rejected.
	if _, err := svc.RetryParse(context.Background(), "admin", initRes.Asset.ID); !errors.Is(err, shared.ErrConflict) {
		t.Fatalf("expected conflict for non-uploaded asset, got %v", err)
	}
}
