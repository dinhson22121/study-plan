package application

import (
	"context"
	"time"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// issueTokenPair mints an access+refresh pair and records the refresh jti so it
// can later be rotated or revoked.
func (s *Service) issueTokenPair(ctx context.Context, userID string, role authdomain.Role) (authdomain.TokenPair, error) {
	var zero authdomain.TokenPair

	access, expUnix, err := s.tokens.IssueAccess(userID, role)
	if err != nil {
		return zero, domain.ErrInternal.WithCause(err)
	}
	refresh, jti, err := s.tokens.IssueRefresh(userID)
	if err != nil {
		return zero, domain.ErrInternal.WithCause(err)
	}
	if err := s.refresh.Save(ctx, userID, jti); err != nil {
		return zero, domain.ErrInternal.WithCause(err)
	}
	return authdomain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    time.Unix(expUnix, 0).UTC(),
	}, nil
}
