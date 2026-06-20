// Package application contains the auth use cases: register, login, refresh,
// logout, and access-token validation. It orchestrates domain ports and never
// touches transport or storage details directly.
package application

import (
	"time"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
)

// Service implements the auth use cases over its domain ports.
type Service struct {
	repo    authdomain.CredentialRepository
	hasher  authdomain.Hasher
	tokens  authdomain.TokenService
	refresh authdomain.RefreshStore
	bus     *eventbus.Bus
	now     func() time.Time
}

// Option customizes a Service (used by tests to inject a fixed clock).
type Option func(*Service)

// WithClock overrides the time source.
func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

// NewService wires the auth use-case service.
func NewService(
	repo authdomain.CredentialRepository,
	hasher authdomain.Hasher,
	tokens authdomain.TokenService,
	refresh authdomain.RefreshStore,
	bus *eventbus.Bus,
	opts ...Option,
) *Service {
	s := &Service{repo: repo, hasher: hasher, tokens: tokens, refresh: refresh, bus: bus, now: time.Now}
	for _, o := range opts {
		o(s)
	}
	return s
}
