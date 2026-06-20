package studyplanhttp

import (
	"github.com/gin-gonic/gin"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/internal/studyplan/application"
)

type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/studyplans", middleware.Auth(h.validate))
	g.POST("/generate", h.generate)
	g.GET("", h.list)
	g.GET("/:id", h.get)
}

type generateRequest struct {
	SubjectID string `json:"subject_id" binding:"required"`
}

func (h *Handler) generate(c *gin.Context) {
	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	plan, err := h.svc.GeneratePlan(c.Request.Context(), middleware.UserIDFrom(c), req.SubjectID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, plan)
}

func (h *Handler) list(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, plans)
}

func (h *Handler) get(c *gin.Context) {
	plan, err := h.svc.GetPlan(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	if plan.UserID != middleware.UserIDFrom(c) {
		httpx.Fail(c, shared.ErrForbidden)
		return
	}
	httpx.OK(c, plan)
}
