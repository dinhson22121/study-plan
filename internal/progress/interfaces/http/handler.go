// Package progresshttp exposes progress endpoints over HTTP (Gin).
package progresshttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/progress/application"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

// Handler adapts HTTP requests to the progress service.
type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

// NewHandler builds the handler.
func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// Routes mounts the progress endpoints under /progress.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/progress", middleware.Auth(h.validate))
	g.GET("", h.overview)
	g.GET("/topics", h.topics)
}

func (h *Handler) overview(c *gin.Context) {
	ov, err := h.svc.GetOverview(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, ov)
}

func (h *Handler) topics(c *gin.Context) {
	ov, err := h.svc.GetOverview(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, ov.Topics)
}
