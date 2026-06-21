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
