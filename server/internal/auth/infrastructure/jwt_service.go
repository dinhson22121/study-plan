package infrastructure

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/shared/domain"
)

type JWTConfig struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

type JWTService struct {
	cfg JWTConfig
	now func() time.Time
}

func NewJWTService(cfg JWTConfig) *JWTService {
	return &JWTService{cfg: cfg, now: time.Now}
}

type accessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (s *JWTService) IssueAccess(userID string, role authdomain.Role) (string, int64, error) {
	now := s.now()
	exp := now.Add(s.cfg.AccessTTL)
	claims := accessClaims{
		Role: string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
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

func (s *JWTService) ParseAccess(token string) (*authdomain.Claims, error) {
	var claims accessClaims
	if err := s.parse(token, &claims); err != nil {
		return nil, err
	}
	out := &authdomain.Claims{UserID: claims.Subject, Role: authdomain.Role(claims.Role), ID: claims.ID}
	if claims.ExpiresAt != nil {
		out.ExpiresAt = claims.ExpiresAt.Time
	}
	return out, nil
}

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
