package domain

import "context"

type CredentialRepository interface {
	Create(ctx context.Context, c *UserCredential) error
	FindByEmail(ctx context.Context, email string) (*UserCredential, error)
	FindByUserID(ctx context.Context, userID string) (*UserCredential, error)
}

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenService interface {
	IssueAccess(userID string, role Role) (token string, expiresAt int64, err error)
	IssueRefresh(userID string) (token string, jti string, err error)
	ParseAccess(token string) (*Claims, error)
	ParseRefresh(token string) (*RefreshClaims, error)
}

type RefreshStore interface {
	Save(ctx context.Context, userID, jti string) error
	Exists(ctx context.Context, userID, jti string) (bool, error)
	Delete(ctx context.Context, userID, jti string) error
}
