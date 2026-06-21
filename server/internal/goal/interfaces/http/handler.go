package goalhttp

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/goal/application"
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
	g := rg.Group("/goals", middleware.Auth(h.validate))
	g.PUT("", h.setGoal)
	g.GET("", h.getGoal)
}

type subjectTargetRequest struct {
	SubjectID    string  `json:"subject_id" binding:"required"`
	CurrentScore float64 `json:"current_score"`
	TargetScore  float64 `json:"target_score"`
}

type setGoalRequest struct {
	TargetUniversity string                 `json:"target_university" binding:"required"`
	TargetMajor      string                 `json:"target_major"`
	TargetDate       time.Time              `json:"target_date" binding:"required"`
	HoursPerDay      int                    `json:"hours_per_day" binding:"required"`
	DaysPerWeek      int                    `json:"days_per_week" binding:"required"`
	Subjects         []subjectTargetRequest `json:"subjects" binding:"required,min=1"`
}

func (h *Handler) setGoal(c *gin.Context) {
	var req setGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	subjects := make([]application.SubjectTargetInput, 0, len(req.Subjects))
	for _, s := range req.Subjects {
		subjects = append(subjects, application.SubjectTargetInput{
			SubjectID: s.SubjectID, CurrentScore: s.CurrentScore, TargetScore: s.TargetScore,
		})
	}
	goal, err := h.svc.SetGoal(c.Request.Context(), application.SetGoalInput{
		UserID:           middleware.UserIDFrom(c),
		TargetUniversity: req.TargetUniversity,
		TargetMajor:      req.TargetMajor,
		TargetDate:       req.TargetDate,
		HoursPerDay:      req.HoursPerDay,
		DaysPerWeek:      req.DaysPerWeek,
		Subjects:         subjects,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, goal)
}

func (h *Handler) getGoal(c *gin.Context) {
	goal, err := h.svc.GetGoal(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, goal)
}
