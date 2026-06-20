package domain

import (
	"path"
	"regexp"
	"strings"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const AllowedUploadContentType = "application/pdf"

type AssetStatus string

type AssetEntityType string

const (
	AssetPending  AssetStatus = "PENDING"
	AssetUploaded AssetStatus = "UPLOADED"
	AssetVerified AssetStatus = "VERIFIED"
	AssetDeleted  AssetStatus = "DELETED"
	AssetFailed   AssetStatus = "FAILED"

	EntityQuestion   AssetEntityType = "QUESTION"
	EntityExam       AssetEntityType = "EXAM"
	EntityContent    AssetEntityType = "CONTENT"
	EntityAttachment AssetEntityType = "ATTACHMENT"
)

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
	EntityType       AssetEntityType
	EntityID         string
	StorageProvider  string
	CreatedAt        time.Time
	VerifiedAt       *time.Time
	DeletedAt        *time.Time
}

func NewPendingAsset(id, objectKey, bucket, filename, contentType, uploadedBy, checksumSHA256 string, now time.Time) (*UploadedAsset, error) {
	if strings.TrimSpace(filename) == "" {
		return nil, shared.ErrValidation.WithMessage("filename is required")
	}
	if contentType != AllowedUploadContentType {
		return nil, shared.ErrValidation.WithMessage("only application/pdf is accepted")
	}
	if uploadedBy == "" {
		return nil, shared.ErrValidation.WithMessage("uploader id is required")
	}
	checksumSHA256 = strings.ToLower(strings.TrimSpace(checksumSHA256))
	if checksumSHA256 != "" && !checksumHex.MatchString(checksumSHA256) {
		return nil, shared.ErrValidation.WithMessage("checksum_sha256 must be a 64-character hex string")
	}
	return &UploadedAsset{
		ID: id, ObjectKey: objectKey, BucketName: bucket,
		OriginalFilename: filename, ContentType: contentType,
		ChecksumSHA256: checksumSHA256,
		Status:         AssetPending, UploadedBy: uploadedBy,
		StorageProvider: "S3", CreatedAt: now,
	}, nil
}

var unsafeKeyChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
var checksumHex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

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

func (t AssetEntityType) Valid() bool {
	switch t {
	case EntityQuestion, EntityExam, EntityContent, EntityAttachment:
		return true
	default:
		return false
	}
}
