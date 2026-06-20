// Package quizhttp exposes the quiz flow over HTTP (Gin).
package quizhttp

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/quiz/application"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

// Handler adapts HTTP requests to the quiz service.
type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

// NewHandler builds the handler.
func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// Routes mounts the quiz endpoints under /quizzes.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/quizzes", middleware.Auth(h.validate))
	g.POST("", h.start)
	g.POST("/:id/submit", h.submit)
	g.GET("/:id", h.getResult)
	g.GET("", h.list)
}

type startRequest struct {
	TopicID      string `json:"topic_id" binding:"required"`
	NumQuestions int    `json:"num_questions"`
}

func (h *Handler) start(c *gin.Context) {
	var req startRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	session, err := h.svc.StartQuiz(c.Request.Context(), middleware.UserIDFrom(c), req.TopicID, req.NumQuestions)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, gin.H{
		"id":           session.ID,
		"topic_id":     session.TopicID,
		"question_ids": session.QuestionIDs,
	})
}

type submitRequest struct {
	Answers []struct {
		QuestionID string `json:"question_id" binding:"required"`
		OptionID   string `json:"option_id"`
	} `json:"answers" binding:"required,min=1"`
}

func (h *Handler) submit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	answers := make([]application.AnswerInput, 0, len(req.Answers))
	for _, a := range req.Answers {
		answers = append(answers, application.AnswerInput{QuestionID: a.QuestionID, OptionID: a.OptionID})
	}
	result, err := h.svc.SubmitQuiz(c.Request.Context(), c.Param("id"), middleware.UserIDFrom(c), answers)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, result)
}

func (h *Handler) getResult(c *gin.Context) {
	result, err := h.svc.GetResult(c.Request.Context(), c.Param("id"), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, result)
}

func (h *Handler) list(c *gin.Context) {
	results, err := h.svc.ListResults(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, results)
}
