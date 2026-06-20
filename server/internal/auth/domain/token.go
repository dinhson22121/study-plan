package domain

import "time"

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Claims struct {
	UserID string
	Role   Role
}

type RefreshClaims struct {
	UserID string
	ID     string
}
