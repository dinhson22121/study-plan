package application

import (
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

func (s *Service) ValidateAccessToken(token string) (*middleware.Claims, error) {
	claims, err := s.tokens.ParseAccess(token)
	if err != nil {
		return nil, domain.ErrUnauthorized.WithCause(err)
	}
	return &middleware.Claims{UserID: claims.UserID, Role: string(claims.Role)}, nil
}
