// Package curriculumhttp exposes the curriculum catalog over HTTP (Gin). Reads
// require authentication; writes require the ADMIN role.
package curriculumhttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/curriculum/application"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

// Handler adapts HTTP requests to the curriculum service.
type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

// NewHandler builds the handler.
func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// Routes mounts the curriculum endpoints under /curriculum.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	auth := middleware.Auth(h.validate)
	admin := middleware.RequireRole(middleware.RoleAdmin)

	g := rg.Group("/curriculum", auth)
	g.GET("/subjects", h.listSubjects)
	g.POST("/subjects", admin, h.createSubject)
	g.GET("/subjects/:id/chapters", h.listChapters)
	g.POST("/subjects/:id/chapters", admin, h.createChapter)
	g.GET("/chapters/:id/topics", h.listTopics)
	g.POST("/chapters/:id/topics", admin, h.createTopic)
	g.GET("/topics/:id", h.getTopic)
}

type createSubjectRequest struct {
	Code       string `json:"code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	GradeLevel int    `json:"grade_level" binding:"required"`
}

func (h *Handler) createSubject(c *gin.Context) {
	var req createSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	subject, err := h.svc.CreateSubject(c.Request.Context(), req.Code, req.Name, req.GradeLevel)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, subject)
}

func (h *Handler) listSubjects(c *gin.Context) {
	subjects, err := h.svc.ListSubjects(c.Request.Context())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, subjects)
}

type createChildRequest struct {
	Title      string `json:"title" binding:"required"`
	OrderIndex int    `json:"order_index"`
}

func (h *Handler) createChapter(c *gin.Context) {
	var req createChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	chapter, err := h.svc.CreateChapter(c.Request.Context(), c.Param("id"), req.Title, req.OrderIndex)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, chapter)
}

func (h *Handler) listChapters(c *gin.Context) {
	chapters, err := h.svc.ListChapters(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, chapters)
}

func (h *Handler) createTopic(c *gin.Context) {
	var req createChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	topic, err := h.svc.CreateTopic(c.Request.Context(), c.Param("id"), req.Title, req.OrderIndex)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, topic)
}

func (h *Handler) listTopics(c *gin.Context) {
	topics, err := h.svc.ListTopics(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, topics)
}

func (h *Handler) getTopic(c *gin.Context) {
	topic, err := h.svc.GetTopic(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, topic)
}
