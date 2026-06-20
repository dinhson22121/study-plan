package domain

import (
	"path"
	"regexp"
	"strings"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// AllowedUploadContentType is the only file type accepted in the MVP.
const AllowedUploadContentType = "application/pdf"

// AssetStatus is the lifecycle state of an uploaded asset.
type AssetStatus string

const (
	AssetPending  AssetStatus = "PENDING"  // session created, awaiting upload
	AssetUploaded AssetStatus = "UPLOADED" // object verified on storage
	AssetDeleted  AssetStatus = "DELETED"  // soft-deleted
	AssetFailed   AssetStatus = "FAILED"
)

// UploadedAsset is an admin-uploaded file's metadata; the bytes live in S3.
type UploadedAsset struct {
	ID               string
	ObjectKey        string
	BucketName       string
	OriginalFilename string
	ContentType      string
	FileSize         int64
	ChecksumSHA256   string
	Status           AssetStatus
	UploadedBy       string
	StorageProvider  string
	CreatedAt        time.Time
	VerifiedAt       *time.Time
	DeletedAt        *time.Time
}

// NewPendingAsset validates input and builds a PENDING asset for an upload
// session. Size limits are enforced in the application layer (config-driven).
func NewPendingAsset(id, objectKey, bucket, filename, contentType string, uploadedBy string, now time.Time) (*UploadedAsset, error) {
	if strings.TrimSpace(filename) == "" {
		return nil, shared.ErrValidation.WithMessage("filename is required")
	}
	if contentType != AllowedUploadContentType {
		return nil, shared.ErrValidation.WithMessage("only application/pdf is accepted")
	}
	if uploadedBy == "" {
		return nil, shared.ErrValidation.WithMessage("uploader id is required")
	}
	return &UploadedAsset{
		ID: id, ObjectKey: objectKey, BucketName: bucket,
		OriginalFilename: filename, ContentType: contentType,
		Status: AssetPending, UploadedBy: uploadedBy,
		StorageProvider: "S3", CreatedAt: now,
	}, nil
}

var unsafeKeyChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// BuildObjectKey produces a collision-resistant, traceable S3 key that does not
// depend solely on the (untrusted) original filename.
func BuildObjectKey(assetID, filename string, now time.Time) string {
	base := sanitizeFilename(path.Base(filename))
	return "exam-assets/" + now.Format("2006/01") + "/" + assetID + "-" + base
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "file.pdf"
	}
	name = unsafeKeyChars.ReplaceAllString(name, "_")
	if len(name) > 120 {
		name = name[len(name)-120:]
	}
	return name
}
