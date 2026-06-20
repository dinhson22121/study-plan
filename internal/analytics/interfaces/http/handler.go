// Package analyticshttp exposes analytics dashboards over HTTP (Gin).
package analyticshttp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/analytics/application"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

// Handler adapts HTTP requests to the analytics service.
type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

// NewHandler builds the handler.
func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// Routes mounts the analytics endpoints under /analytics.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/analytics", middleware.Auth(h.validate))
	g.GET("/me", h.dashboard)
	g.GET("/me/weak-topics", h.weakTopics)
}

func (h *Handler) dashboard(c *gin.Context) {
	dash, err := h.svc.Dashboard(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, dash)
}

func (h *Handler) weakTopics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit <= 0 || limit > 50 {
		limit = 5
	}
	weak, err := h.svc.WeakTopics(c.Request.Context(), middleware.UserIDFrom(c), limit)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, weak)
}
