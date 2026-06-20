// Package s3 wraps the AWS SDK v2 S3 client for the small surface this app needs:
// presigned PUT URLs (admin uploads straight to storage) and HEAD (verify an
// upload completed). It targets any S3-compatible store; MinIO in local dev.
package s3

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Config holds the connection + presign settings.
type Config struct {
	Endpoint     string // empty = real AWS S3
	Region       string
	AccessKey    string
	SecretKey    string
	Bucket       string
	UsePathStyle bool
	PresignTTL   time.Duration
}

// Client is the configured S3 wrapper.
type Client struct {
	api     *awss3.Client
	presign *awss3.PresignClient
	bucket  string
	ttl     time.Duration
}

// PresignedUpload is everything the admin client needs to PUT a file directly.
type PresignedUpload struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

// New builds the client from static credentials. It does not contact S3.
func New(cfg Config) (*Client, error) {
	if cfg.Bucket == "" {
		return nil, errors.New("s3: bucket is required")
	}
	awsCfg := aws.Config{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	}
	api := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})
	ttl := cfg.PresignTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &Client{api: api, presign: awss3.NewPresignClient(api), bucket: cfg.Bucket, ttl: ttl}, nil
}

// Bucket returns the configured bucket name.
func (c *Client) Bucket() string { return c.bucket }

// PresignPut returns a presigned PUT URL for the object key (computed locally).
func (c *Client) PresignPut(ctx context.Context, key, contentType string) (PresignedUpload, error) {
	req, err := c.presign.PresignPutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, awss3.WithPresignExpires(c.ttl))
	if err != nil {
		return PresignedUpload{}, err
	}
	headers := make(map[string]string, len(req.SignedHeader))
	for k, v := range req.SignedHeader {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	return PresignedUpload{
		URL:       req.URL,
		Method:    req.Method,
		Headers:   headers,
		ExpiresAt: time.Now().Add(c.ttl),
	}, nil
}

// Head returns whether the object exists plus its size and content type.
func (c *Client) Head(ctx context.Context, key string) (found bool, size int64, contentType string, err error) {
	out, herr := c.api.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if herr != nil {
		var notFound *types.NotFound
		if errors.As(herr, &notFound) {
			return false, 0, "", nil
		}
		return false, 0, "", herr
	}
	return true, aws.ToInt64(out.ContentLength), aws.ToString(out.ContentType), nil
}
