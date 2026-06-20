package application

import (
	"context"
	"errors"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

type LoginInput struct {
	Email    string
	Password string
}

func (s *Service) Login(ctx context.Context, in LoginInput) (authdomain.TokenPair, error) {
	var zero authdomain.TokenPair

	email := authdomain.NormalizeEmail(in.Email)
	cred, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return zero, domain.ErrUnauthorized.WithMessage("invalid credentials")
		}
		return zero, err
	}

	if err := s.hasher.Compare(cred.PasswordHash, in.Password); err != nil {
		return zero, domain.ErrUnauthorized.WithMessage("invalid credentials")
	}

	return s.issueTokenPair(ctx, cred.UserID, cred.Role)
}
