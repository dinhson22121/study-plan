package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

func TestBuildObjectKey(t *testing.T) {
	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	key := BuildObjectKey("asset-123", "Đề thi Toán 2026.pdf", now)

	if !strings.HasPrefix(key, "exam-assets/2026/01/asset-123-") {
		t.Fatalf("unexpected key prefix: %s", key)
	}
	// Unsafe chars (spaces, accents) must be sanitized out of the key.
	if strings.ContainsAny(key[len("exam-assets/2026/01/"):], " ") {
		t.Fatalf("key should not contain spaces: %s", key)
	}
}

func TestNewPendingAsset_Validation(t *testing.T) {
	now := time.Unix(0, 0)
	if _, err := NewPendingAsset("id", "k", "b", "", "application/pdf", "admin", now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty filename")
	}
	if _, err := NewPendingAsset("id", "k", "b", "f.png", "image/png", "admin", now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for non-pdf content type")
	}
	if _, err := NewPendingAsset("id", "k", "b", "f.pdf", "application/pdf", "", now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("expected error for empty uploader")
	}
	a, err := NewPendingAsset("id", "k", "b", "f.pdf", "application/pdf", "admin", now)
	if err != nil || a.Status != AssetPending || a.StorageProvider != "S3" {
		t.Fatalf("expected valid pending asset, got %+v / %v", a, err)
	}
}
