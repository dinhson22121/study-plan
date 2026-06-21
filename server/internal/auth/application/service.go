package application

import (
	"time"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
)

type Service struct {
	repo      authdomain.CredentialRepository
	hasher    authdomain.Hasher
	tokens    authdomain.TokenService
	refresh   authdomain.RefreshStore
	blocklist authdomain.TokenBlocklist
	bus       *eventbus.Bus
	now       func() time.Time
}

type Option func(*Service)

func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

func WithBlocklist(bl authdomain.TokenBlocklist) Option {
	return func(s *Service) { s.blocklist = bl }
}

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
