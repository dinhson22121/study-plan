// Package userhttp exposes user profile endpoints over HTTP (Gin).
package userhttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/internal/user/application"
)

// Handler adapts HTTP requests to the user application service.
type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

// NewHandler builds the user HTTP handler.
func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// Routes mounts the protected profile endpoints under the given group.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/users", middleware.Auth(h.validate))
	g.GET("/me", h.getMe)
	g.PUT("/me", h.updateMe)
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
}

func (h *Handler) getMe(c *gin.Context) {
	profile, err := h.svc.GetProfile(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, profile)
}

func (h *Handler) updateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, domain.ErrValidation.WithCause(err))
		return
	}
	profile, err := h.svc.UpdateDisplayName(c.Request.Context(), middleware.UserIDFrom(c), req.DisplayName)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, profile)
}
