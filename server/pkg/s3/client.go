package s3

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Endpoint     string
	Region       string
	AccessKey    string
	SecretKey    string
	Bucket       string
	UsePathStyle bool
	PresignTTL   time.Duration
}

type Client struct {
	api     *awss3.Client
	presign *awss3.PresignClient
	bucket  string
	ttl     time.Duration
}

type PresignedUpload struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

func New(cfg Config) (*Client, error) {
	if cfg.Bucket == "" {
		return nil, errors.New("s3: bucket is required")
	}
	if (cfg.AccessKey == "") != (cfg.SecretKey == "") {
		return nil, errors.New("s3: access_key and secret_key must both be set or both be empty")
	}
	loadOpts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.Region)}
	if cfg.AccessKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOpts...)
	if err != nil {
		return nil, err
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

func (c *Client) Bucket() string { return c.bucket }

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

func (c *Client) ReadAll(ctx context.Context, key string) ([]byte, error) {
	out, err := c.api.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
