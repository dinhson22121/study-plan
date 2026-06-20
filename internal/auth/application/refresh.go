package application

import (
	"context"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

func (s *Service) Refresh(ctx context.Context, refreshToken string) (authdomain.TokenPair, error) {
	var zero authdomain.TokenPair

	claims, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return zero, domain.ErrUnauthorized.WithMessage("invalid refresh token")
	}

	active, err := s.refresh.Exists(ctx, claims.UserID, claims.ID)
	if err != nil {
		return zero, domain.ErrInternal.WithCause(err)
	}
	if !active {
		return zero, domain.ErrUnauthorized.WithMessage("refresh token revoked")
	}

	cred, err := s.repo.FindByUserID(ctx, claims.UserID)
	if err != nil {
		return zero, err
	}

	if err := s.refresh.Delete(ctx, claims.UserID, claims.ID); err != nil {
		return zero, domain.ErrInternal.WithCause(err)
	}
	return s.issueTokenPair(ctx, cred.UserID, cred.Role)
}
