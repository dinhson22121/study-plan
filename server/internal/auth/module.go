package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/auth/application"
	"github.com/son-ngo/edu-app/internal/auth/infrastructure"
	authhttp "github.com/son-ngo/edu-app/internal/auth/interfaces/http"
	"github.com/son-ngo/edu-app/internal/shared/ratelimit"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	svc := NewService(deps)
	deps.AuthValidate = svc.ValidateAccessToken
	if deps.AuthRateLimiter == nil {
		deps.AuthRateLimiter = ratelimit.NewRedisLimiter(
			deps.Redis, deps.Cfg.RateLimit.AuthRequests, deps.Cfg.RateLimit.AuthWindow)
	}
	authhttp.NewHandler(svc, svc.ValidateAccessToken, deps.AuthRateLimiter).Routes(rg)
}

func NewService(deps *app.Deps) *application.Service {
	repo := infrastructure.NewPgCredentialRepo(deps.DB)
	hasher := infrastructure.NewBcryptHasher(0)
	tokens := infrastructure.NewJWTService(infrastructure.JWTConfig{
		Secret:     []byte(deps.Cfg.JWT.Secret),
		AccessTTL:  deps.Cfg.JWT.AccessTTL,
		RefreshTTL: deps.Cfg.JWT.RefreshTTL,
		Issuer:     deps.Cfg.JWT.Issuer,
	})
	refreshStore := infrastructure.NewRedisRefreshStore(deps.Redis, deps.Cfg.JWT.RefreshTTL)
	blocklist := infrastructure.NewRedisBlocklist(deps.Redis)
	return application.NewService(repo, hasher, tokens, refreshStore, deps.Bus,
		application.WithBlocklist(blocklist))
}
