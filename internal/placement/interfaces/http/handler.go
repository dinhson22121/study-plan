package placementhttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/placement/application"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/placement", middleware.Auth(h.validate))
	g.POST("/tests", h.startTest)
	g.POST("/tests/:id/submit", h.submitTest)
	g.GET("/results", h.listResults)
}

type startTestRequest struct {
	SubjectID    string `json:"subject_id" binding:"required"`
	NumQuestions int    `json:"num_questions"`
}

func (h *Handler) startTest(c *gin.Context) {
	var req startTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	test, err := h.svc.StartTest(c.Request.Context(), middleware.UserIDFrom(c), req.SubjectID, req.NumQuestions)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.Created(c, gin.H{
		"id":           test.ID,
		"subject_id":   test.SubjectID,
		"question_ids": test.QuestionIDs,
	})
}

type submitRequest struct {
	Answers []struct {
		QuestionID string `json:"question_id" binding:"required"`
		OptionID   string `json:"option_id" binding:"required"`
	} `json:"answers" binding:"required,min=1"`
}

func (h *Handler) submitTest(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	answers := make([]application.AnswerInput, 0, len(req.Answers))
	for _, a := range req.Answers {
		answers = append(answers, application.AnswerInput{QuestionID: a.QuestionID, OptionID: a.OptionID})
	}
	result, err := h.svc.SubmitTest(c.Request.Context(), c.Param("id"), middleware.UserIDFrom(c), answers)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, result)
}

func (h *Handler) listResults(c *gin.Context) {
	results, err := h.svc.ListResults(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, results)
}
