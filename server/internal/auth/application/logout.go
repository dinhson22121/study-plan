package application

import (
	"context"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return nil
	}
	if err := s.refresh.Delete(ctx, claims.UserID, claims.ID); err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *Service) RevokeAccessToken(ctx context.Context, accessToken string) error {
	if s.blocklist == nil {
		return nil
	}
	claims, err := s.tokens.ParseAccess(accessToken)
	if err != nil || claims.ID == "" {
		return nil
	}
	ttl := claims.ExpiresAt.Sub(s.now())
	if ttl <= 0 {
		return nil
	}
	if err := s.blocklist.Revoke(ctx, claims.ID, ttl); err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}
