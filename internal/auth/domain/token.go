package domain

import "time"

// TokenPair is the credential set returned on login/refresh.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims is the validated identity carried by an access token.
type Claims struct {
	UserID string
	Role   Role
}

// RefreshClaims is the validated identity carried by a refresh token. The ID is
// the token's unique identifier (jti), used to revoke it from the store.
type RefreshClaims struct {
	UserID string
	ID     string
}
