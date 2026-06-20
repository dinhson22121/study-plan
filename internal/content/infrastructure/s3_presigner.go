package infrastructure

import (
	"context"

	"github.com/son-ngo/edu-app/internal/content/domain"
	s3pkg "github.com/son-ngo/edu-app/pkg/s3"
)

// S3Storage adapts pkg/s3.Client to the content ObjectStorage port.
type S3Storage struct {
	client *s3pkg.Client
}

// NewS3Storage builds the adapter.
func NewS3Storage(client *s3pkg.Client) *S3Storage { return &S3Storage{client: client} }

// Bucket returns the configured bucket name.
func (a *S3Storage) Bucket() string { return a.client.Bucket() }

// PresignPut returns presigned PUT instructions in domain shape.
func (a *S3Storage) PresignPut(ctx context.Context, key, contentType string) (domain.PresignedUpload, error) {
	up, err := a.client.PresignPut(ctx, key, contentType)
	if err != nil {
		return domain.PresignedUpload{}, err
	}
	return domain.PresignedUpload{
		URL: up.URL, Method: up.Method, Headers: up.Headers, ExpiresAt: up.ExpiresAt,
	}, nil
}

// Head reports object existence/size/content-type.
func (a *S3Storage) Head(ctx context.Context, key string) (bool, int64, string, error) {
	return a.client.Head(ctx, key)
}

var _ domain.ObjectStorage = (*S3Storage)(nil)
