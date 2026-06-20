package domain

import (
	"strings"
	"time"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

// DeviceToken is a user's FCM registration token for one device.
type DeviceToken struct {
	ID        string
	UserID    string
	Token     string
	Platform  Platform
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewDeviceToken validates and constructs an active device token.
func NewDeviceToken(id, userID, token string, platform Platform, now time.Time) (*DeviceToken, error) {
	if userID == "" {
		return nil, shared.ErrValidation.WithMessage("user id required")
	}
	if strings.TrimSpace(token) == "" {
		return nil, shared.ErrValidation.WithMessage("device token required")
	}
	if !platform.Valid() {
		return nil, shared.ErrValidation.WithMessage("platform must be android or ios")
	}
	return &DeviceToken{
		ID:        id,
		UserID:    userID,
		Token:     token,
		Platform:  platform,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
