package authhttp

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/auth/application"
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
	limiter  middleware.RateLimiter
}

func NewHandler(svc *application.Service, validate middleware.TokenValidator, limiter middleware.RateLimiter) *Handler {
	return &Handler{svc: svc, validate: validate, limiter: limiter}
}

func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	throttle := middleware.RateLimit(h.limiter, "auth")
	g.POST("/register", throttle, h.register)
	g.POST("/login", throttle, h.login)
	g.POST("/refresh", throttle, h.refresh)
	g.POST("/logout", middleware.Auth(h.validate), h.logout)
}

type credentialsRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) register(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, domain.ErrValidation.WithCause(err))
		return
	}
	pair, err := h.svc.Register(c.Request.Context(), application.RegisterInput{Email: req.Email, Password: req.Password})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, pair)
}

func (h *Handler) login(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, domain.ErrValidation.WithCause(err))
		return
	}
	pair, err := h.svc.Login(c.Request.Context(), application.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, pair)
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, domain.ErrValidation.WithCause(err))
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, pair)
}

func (h *Handler) logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, domain.ErrValidation.WithCause(err))
		return
	}
	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		httpx.Fail(c, err)
		return
	}
	if access := bearerToken(c); access != "" {
		if err := h.svc.RevokeAccessToken(c.Request.Context(), access); err != nil {
			httpx.Fail(c, err)
			return
		}
	}
	httpx.OK(c, gin.H{"message": "logged out"})
}

func bearerToken(c *gin.Context) string {
	parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1]) // already validated by the Auth middleware
}
