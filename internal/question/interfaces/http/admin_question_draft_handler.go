package questionhttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/question/application"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

type AdminDraftHandler struct {
	svc      *application.DraftService
	validate middleware.TokenValidator
}

func NewAdminDraftHandler(svc *application.DraftService, validate middleware.TokenValidator) *AdminDraftHandler {
	return &AdminDraftHandler{svc: svc, validate: validate}
}

func (h *AdminDraftHandler) Routes(rg *gin.RouterGroup) {
	admin := []gin.HandlerFunc{middleware.Auth(h.validate), middleware.RequireRole(middleware.RoleAdmin)}

	rg.GET("/admin/uploads/:id/draft-questions", append(admin, h.listByAsset)...)
	rg.POST("/admin/uploads/:id/publish", append(admin, h.publishByAsset)...)

	d := rg.Group("/admin/question-drafts", admin...)
	d.PUT("/:id", h.updateDraft)
	d.PUT("/:id/options/:optionId", h.updateOption)
	d.POST("/:id/publish", h.publish)
}

func (h *AdminDraftHandler) listByAsset(c *gin.Context) {
	drafts, err := h.svc.ListByAsset(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, drafts)
}

type updateDraftRequest struct {
	Stem        string `json:"stem" binding:"required"`
	Explanation string `json:"explanation"`
}

func (h *AdminDraftHandler) updateDraft(c *gin.Context) {
	var req updateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	if err := h.svc.UpdateDraft(c.Request.Context(), c.Param("id"), req.Stem, req.Explanation); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"message": "draft updated"})
}

type updateOptionRequest struct {
	Text      string `json:"text" binding:"required"`
	IsCorrect bool   `json:"is_correct"`
}

func (h *AdminDraftHandler) updateOption(c *gin.Context) {
	var req updateOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	if err := h.svc.UpdateOption(c.Request.Context(), c.Param("optionId"), req.Text, req.IsCorrect); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"message": "option updated"})
}

type publishRequest struct {
	TopicID    string `json:"topic_id" binding:"required"`
	Difficulty string `json:"difficulty" binding:"required"`
}

func (h *AdminDraftHandler) publish(c *gin.Context) {
	var req publishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	q, err := h.svc.Publish(c.Request.Context(), application.PublishInput{
		DraftID: c.Param("id"), TopicID: req.TopicID, Difficulty: req.Difficulty,
		ReviewedBy: middleware.UserIDFrom(c),
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, q)
}

func (h *AdminDraftHandler) publishByAsset(c *gin.Context) {
	var req publishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	questions, err := h.svc.PublishByAsset(c.Request.Context(), application.PublishByAssetInput{
		AssetID: c.Param("id"), TopicID: req.TopicID, Difficulty: req.Difficulty, ReviewedBy: middleware.UserIDFrom(c),
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, questions)
}
