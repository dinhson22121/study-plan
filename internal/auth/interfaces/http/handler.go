// Package authhttp exposes the auth use cases over HTTP (Gin).
package authhttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/auth/application"
	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
)

// Handler adapts HTTP requests to the auth application service.
type Handler struct {
	svc *application.Service
}

// NewHandler builds the auth HTTP handler.
func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

// Routes registers the auth endpoints under the given group.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.POST("/register", h.register)
	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/logout", h.logout)
}

type credentialsRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
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
	httpx.OK(c, gin.H{"message": "logged out"})
}
