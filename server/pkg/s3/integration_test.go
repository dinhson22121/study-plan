//go:build integration

package s3

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
)

func testConfig(t *testing.T) Config {
	t.Helper()

	endpoint := getenvAny("EDU_TEST_S3_ENDPOINT", "EDU_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("EDU_TEST_S3_ENDPOINT / EDU_S3_ENDPOINT not set")
	}

	bucket := getenvAny("EDU_TEST_S3_BUCKET", "EDU_S3_BUCKET")
	if bucket == "" {
		t.Skip("EDU_TEST_S3_BUCKET / EDU_S3_BUCKET not set")
	}

	usePathStyle := true
	if v := getenvAny("EDU_TEST_S3_USE_PATH_STYLE", "EDU_S3_USE_PATH_STYLE"); v == "false" {
		usePathStyle = false
	}

	return Config{
		Endpoint:     endpoint,
		Region:       getenvDefault(getenvAny("EDU_TEST_S3_REGION", "EDU_S3_REGION"), "us-east-1"),
		AccessKey:    getenvAny("EDU_TEST_S3_ACCESS_KEY", "EDU_S3_ACCESS_KEY"),
		SecretKey:    getenvAny("EDU_TEST_S3_SECRET_KEY", "EDU_S3_SECRET_KEY"),
		Bucket:       bucket,
		UsePathStyle: usePathStyle,
	}
}

func getenvAny(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

func getenvDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func TestClient_PresignPutHeadAndReadAll(t *testing.T) {
	cfg := testConfig(t)

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	key := "integration/" + uuid.NewString() + ".pdf"
	body := []byte("%PDF-1.4\nMinIO integration test\n")

	ctx := context.Background()
	upload, err := client.PresignPut(ctx, key, "application/pdf")
	if err != nil {
		t.Fatalf("presign put: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, upload.Method, upload.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range upload.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/pdf")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload via presigned url: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		t.Fatalf("unexpected upload status: %s", resp.Status)
	}

	found, size, contentType, err := client.Head(ctx, key)
	if err != nil {
		t.Fatalf("head object: %v", err)
	}
	if !found {
		t.Fatalf("expected object to exist after upload")
	}
	if size != int64(len(body)) {
		t.Fatalf("unexpected object size: got %d want %d", size, len(body))
	}
	if contentType == "" {
		t.Fatalf("expected content type to be present")
	}

	got, err := client.ReadAll(ctx, key)
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("object body mismatch")
	}
}
