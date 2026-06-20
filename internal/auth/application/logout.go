package application

import (
	"context"

	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// Logout revokes a refresh token by deleting its jti from the store. Parsing an
// invalid token is treated as already-logged-out (idempotent success).
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return nil // nothing to revoke
	}
	if err := s.refresh.Delete(ctx, claims.UserID, claims.ID); err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}
