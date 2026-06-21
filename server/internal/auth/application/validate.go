package application

import (
	"context"

	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

func (s *Service) ValidateAccessToken(ctx context.Context, token string) (*middleware.Claims, error) {
	claims, err := s.tokens.ParseAccess(token)
	if err != nil {
		return nil, domain.ErrUnauthorized.WithCause(err)
	}
	if s.blocklist != nil && claims.ID != "" {
		revoked, err := s.blocklist.IsRevoked(ctx, claims.ID)
		if err != nil {
			return nil, domain.ErrInternal.WithCause(err)
		}
		if revoked {
			return nil, domain.ErrUnauthorized.WithMessage("token revoked")
		}
	}
	return &middleware.Claims{UserID: claims.UserID, Role: string(claims.Role)}, nil
}
