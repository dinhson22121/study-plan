package notifhttp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/notification/application"
	"github.com/son-ngo/edu-app/internal/notification/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

type Handler struct {
	mgr      *application.Manager
	validate middleware.TokenValidator
}

func NewHandler(mgr *application.Manager, validate middleware.TokenValidator) *Handler {
	return &Handler{mgr: mgr, validate: validate}
}

func (h *Handler) Routes(rg *gin.RouterGroup) {
	auth := middleware.Auth(h.validate)

	devices := rg.Group("/devices", auth)
	devices.POST("/token", h.registerToken)
	devices.DELETE("/token", h.deleteToken)

	notif := rg.Group("/notifications", auth)
	notif.GET("/preferences", h.listPreferences)
	notif.PUT("/preferences/:type", h.setPreference)
	notif.GET("/history", h.history)

	admin := rg.Group("/admin/notifications", auth, middleware.RequireRole(middleware.RoleAdmin))
	admin.POST("/broadcast", h.broadcast)
}

type tokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=android ios"`
}

func (h *Handler) registerToken(c *gin.Context) {
	var req tokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	if err := h.mgr.RegisterDeviceToken(c.Request.Context(), middleware.UserIDFrom(c), req.Token, req.Platform); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, gin.H{"message": "token registered"})
}

type deleteTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *Handler) deleteToken(c *gin.Context) {
	var req deleteTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	if err := h.mgr.DeleteDeviceToken(c.Request.Context(), middleware.UserIDFrom(c), req.Token); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"message": "token deleted"})
}

func (h *Handler) listPreferences(c *gin.Context) {
	prefs, err := h.mgr.ListPreferences(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, prefs)
}

type setPreferenceRequest struct {
	Enabled *bool `json:"enabled" binding:"required"`
}

func (h *Handler) setPreference(c *gin.Context) {
	var req setPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	if err := h.mgr.SetPreference(c.Request.Context(), middleware.UserIDFrom(c), c.Param("type"), *req.Enabled); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"message": "preference updated"})
}

func (h *Handler) history(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	logs, total, err := h.mgr.GetHistory(c.Request.Context(), middleware.UserIDFrom(c), limit, offset)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.List(c, logs, httpx.Meta{Total: total, Page: offset/maxInt(limit, 1) + 1, Limit: limit})
}

type broadcastRequest struct {
	Type         string            `json:"type" binding:"required"`
	TemplateCode string            `json:"template_code" binding:"required"`
	Variables    map[string]string `json:"variables"`
}

func (h *Handler) broadcast(c *gin.Context) {
	var req broadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	nt, err := domain.ParseType(req.Type)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	count, err := h.mgr.Broadcast(c.Request.Context(), application.BroadcastInput{
		Type: nt, TemplateCode: req.TemplateCode, Variables: req.Variables,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"enqueued": count})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
