// Package contenthttp exposes lessons/content over HTTP (Gin). Authoring is
// ADMIN-only; reads are authenticated.
package contenthttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/content/application"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

// Handler adapts HTTP requests to the content service.
type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

// NewHandler builds the handler.
func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// Routes mounts the content endpoints.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	auth := middleware.Auth(h.validate)
	g := rg.Group("", auth)
	g.GET("/topics/:id/lessons", h.listByTopic)
	g.POST("/topics/:id/lessons", middleware.RequireRole(middleware.RoleAdmin), h.create)
	g.GET("/lessons/:id", h.get)
}

type itemRequest struct {
	Kind string `json:"kind" binding:"required"`
	URL  string `json:"url"`
	Body string `json:"body"`
}

type createLessonRequest struct {
	Title      string        `json:"title" binding:"required"`
	OrderIndex int           `json:"order_index"`
	Items      []itemRequest `json:"items"`
}

func (h *Handler) create(c *gin.Context) {
	var req createLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	items := make([]application.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, application.ItemInput{Kind: it.Kind, URL: it.URL, Body: it.Body})
	}
	lesson, err := h.svc.CreateLesson(c.Request.Context(), application.CreateLessonInput{
		TopicID: c.Param("id"), Title: req.Title, OrderIndex: req.OrderIndex, Items: items,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, lesson)
}

func (h *Handler) listByTopic(c *gin.Context) {
	lessons, err := h.svc.ListByTopic(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, lessons)
}

func (h *Handler) get(c *gin.Context) {
	lesson, err := h.svc.GetLesson(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, lesson)
}
