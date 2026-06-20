package questionhttp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/question/application"
	"github.com/son-ngo/edu-app/internal/question/domain"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type Handler struct {
	svc      *application.Service
	validate middleware.TokenValidator
}

func NewHandler(svc *application.Service, validate middleware.TokenValidator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

func (h *Handler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/questions", middleware.Auth(h.validate))
	g.POST("", middleware.RequireRole(middleware.RoleAdmin), h.create)
	g.GET("/:id", h.get)
	g.GET("", h.list)
}

type optionRequest struct {
	Text      string `json:"text" binding:"required"`
	IsCorrect bool   `json:"is_correct"`
}

type createRequest struct {
	TopicID     string          `json:"topic_id" binding:"required"`
	Type        string          `json:"type" binding:"required"`
	Stem        string          `json:"stem" binding:"required"`
	Difficulty  string          `json:"difficulty" binding:"required"`
	Explanation string          `json:"explanation"`
	Options     []optionRequest `json:"options"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	opts := make([]application.OptionInput, 0, len(req.Options))
	for _, o := range req.Options {
		opts = append(opts, application.OptionInput{Text: o.Text, IsCorrect: o.IsCorrect})
	}
	q, err := h.svc.Create(c.Request.Context(), application.CreateInput{
		TopicID: req.TopicID, Type: req.Type, Stem: req.Stem,
		Difficulty: req.Difficulty, Explanation: req.Explanation, Options: opts,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, toResponse(q, true))
}

func (h *Handler) get(c *gin.Context) {
	q, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, toResponse(q, isAdmin(c)))
}

func (h *Handler) list(c *gin.Context) {
	topicID := c.Query("topic_id")
	if topicID == "" {
		httpx.Fail(c, shared.ErrValidation.WithMessage("topic_id is required"))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > maxListLimit {
		limit = defaultListLimit
	}
	questions, err := h.svc.List(c.Request.Context(), topicID, c.Query("difficulty"), limit)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	admin := isAdmin(c)
	out := make([]questionResponse, 0, len(questions))
	for i := range questions {
		out = append(out, toResponse(&questions[i], admin))
	}
	httpx.OK(c, out)
}

func isAdmin(c *gin.Context) bool { return middleware.RoleFrom(c) == middleware.RoleAdmin }

type optionResponse struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	OrderIndex int    `json:"order_index"`
	IsCorrect  *bool  `json:"is_correct,omitempty"`
}

type questionResponse struct {
	ID          string           `json:"id"`
	TopicID     string           `json:"topic_id"`
	Type        string           `json:"type"`
	Stem        string           `json:"stem"`
	Difficulty  string           `json:"difficulty"`
	Explanation string           `json:"explanation,omitempty"`
	Options     []optionResponse `json:"options"`
}

func toResponse(q *domain.Question, includeAnswers bool) questionResponse {
	resp := questionResponse{
		ID: q.ID, TopicID: q.TopicID, Type: string(q.Type),
		Stem: q.Stem, Difficulty: string(q.Difficulty),
	}
	if includeAnswers {
		resp.Explanation = q.Explanation
	}
	for _, o := range q.Options {
		or := optionResponse{ID: o.ID, Text: o.Text, OrderIndex: o.OrderIndex}
		if includeAnswers {
			v := o.IsCorrect
			or.IsCorrect = &v
		}
		resp.Options = append(resp.Options, or)
	}
	return resp
}
