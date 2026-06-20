package infrastructure

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

// JWTConfig configures token issuance.
type JWTConfig struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

// JWTService implements authdomain.TokenService using HMAC-signed JWTs. Access
// tokens carry the role; refresh tokens carry a unique jti for revocation.
type JWTService struct {
	cfg JWTConfig
	now func() time.Time
}

// NewJWTService builds the service.
func NewJWTService(cfg JWTConfig) *JWTService {
	return &JWTService{cfg: cfg, now: time.Now}
}

type accessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// IssueAccess mints a signed access token and returns it with its expiry (unix).
func (s *JWTService) IssueAccess(userID string, role authdomain.Role) (string, int64, error) {
	now := s.now()
	exp := now.Add(s.cfg.AccessTTL)
	claims := accessClaims{
		Role: string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.cfg.Secret)
	if err != nil {
		return "", 0, domain.ErrInternal.WithCause(err)
	}
	return signed, exp.Unix(), nil
}

// IssueRefresh mints a signed refresh token and returns it with its jti.
func (s *JWTService) IssueRefresh(userID string) (string, string, error) {
	now := s.now()
	jti := uuid.NewString()
	claims := jwt.RegisteredClaims{
		ID:        jti,
		Subject:   userID,
		Issuer:    s.cfg.Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTTL)),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.cfg.Secret)
	if err != nil {
		return "", "", domain.ErrInternal.WithCause(err)
	}
	return signed, jti, nil
}

// ParseAccess validates an access token's signature and expiry, returning claims.
func (s *JWTService) ParseAccess(token string) (*authdomain.Claims, error) {
	var claims accessClaims
	if err := s.parse(token, &claims); err != nil {
		return nil, err
	}
	return &authdomain.Claims{UserID: claims.Subject, Role: authdomain.Role(claims.Role)}, nil
}

// ParseRefresh validates a refresh token, returning its user id and jti.
func (s *JWTService) ParseRefresh(token string) (*authdomain.RefreshClaims, error) {
	var claims jwt.RegisteredClaims
	if err := s.parse(token, &claims); err != nil {
		return nil, err
	}
	return &authdomain.RefreshClaims{UserID: claims.Subject, ID: claims.ID}, nil
}

func (s *JWTService) parse(token string, claims jwt.Claims) error {
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrUnauthorized.WithMessage("unexpected signing method")
		}
		return s.cfg.Secret, nil
	}, jwt.WithIssuer(s.cfg.Issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return domain.ErrUnauthorized.WithCause(err)
	}
	return nil
}
