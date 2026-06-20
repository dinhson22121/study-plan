package infrastructure

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/son-ngo/edu-app/internal/content/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// PgAssetRepository implements domain.AssetRepository over Postgres.
type PgAssetRepository struct {
	db *pgxpool.Pool
}

// NewPgAssetRepository builds the repository.
func NewPgAssetRepository(db *pgxpool.Pool) *PgAssetRepository { return &PgAssetRepository{db: db} }

func (r *PgAssetRepository) Create(ctx context.Context, a *domain.UploadedAsset) error {
	const q = `
		INSERT INTO uploaded_asset
			(id, object_key, bucket_name, original_filename, content_type, file_size,
			 checksum_sha256, status, uploaded_by, storage_provider, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.Exec(ctx, q, a.ID, a.ObjectKey, a.BucketName, a.OriginalFilename, a.ContentType,
		a.FileSize, a.ChecksumSHA256, string(a.Status), a.UploadedBy, a.StorageProvider, a.CreatedAt)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *PgAssetRepository) GetByID(ctx context.Context, id string) (*domain.UploadedAsset, error) {
	const q = `
		SELECT id, object_key, bucket_name, original_filename, content_type, file_size,
		       checksum_sha256, status, uploaded_by, storage_provider, created_at, verified_at, deleted_at
		FROM uploaded_asset WHERE id = $1`
	a, err := scanAsset(r.db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *PgAssetRepository) List(ctx context.Context, f domain.AssetFilter) ([]domain.UploadedAsset, int, error) {
	args := []any{}
	where := "WHERE deleted_at IS NULL"
	if f.Status != "" {
		where += " AND status = $1"
		args = append(args, string(f.Status))
	}

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM uploaded_asset "+where, args...).Scan(&total); err != nil {
		return nil, 0, shared.ErrInternal.WithCause(err)
	}

	q := `SELECT id, object_key, bucket_name, original_filename, content_type, file_size,
	             checksum_sha256, status, uploaded_by, storage_provider, created_at, verified_at, deleted_at
	      FROM uploaded_asset ` + where + " ORDER BY created_at DESC"
	args = append(args, f.Limit, f.Offset)
	q += " LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, shared.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	var out []domain.UploadedAsset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, rows.Err()
}

func (r *PgAssetRepository) MarkUploaded(ctx context.Context, id string, size int64, verifiedAt time.Time) error {
	const q = `UPDATE uploaded_asset SET status = 'UPLOADED', file_size = $2, verified_at = $3 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, size, verifiedAt)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *PgAssetRepository) SoftDelete(ctx context.Context, id string, at time.Time) error {
	const q = `UPDATE uploaded_asset SET status = 'DELETED', deleted_at = $2 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, at)
	if err != nil {
		return shared.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAsset(row rowScanner) (*domain.UploadedAsset, error) {
	var a domain.UploadedAsset
	var status string
	err := row.Scan(&a.ID, &a.ObjectKey, &a.BucketName, &a.OriginalFilename, &a.ContentType, &a.FileSize,
		&a.ChecksumSHA256, &status, &a.UploadedBy, &a.StorageProvider, &a.CreatedAt, &a.VerifiedAt, &a.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound.WithMessage("asset not found")
		}
		return nil, shared.ErrInternal.WithCause(err)
	}
	a.Status = domain.AssetStatus(status)
	return &a, nil
}

var _ domain.AssetRepository = (*PgAssetRepository)(nil)
