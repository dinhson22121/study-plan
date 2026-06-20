package analytics

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/analytics/application"
	"github.com/son-ngo/edu-app/internal/analytics/infrastructure"
	analyticshttp "github.com/son-ngo/edu-app/internal/analytics/interfaces/http"
	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/progress"
	"github.com/son-ngo/edu-app/internal/quiz"
	quizdomain "github.com/son-ngo/edu-app/internal/quiz/domain"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	activityRepo := infrastructure.NewPgActivityRepo(deps.DB)
	progressReader := infrastructure.NewProgressReaderAdapter(progress.NewService(deps))
	quizReader := infrastructure.NewQuizReaderAdapter(quiz.NewService(deps))

	svc := application.NewService(activityRepo, progressReader, quizReader, deps.Log)

	deps.Bus.Subscribe(quizdomain.EventQuizCompleted, svc.HandleQuizCompleted)
	deps.ReengagementSource = svc

	analyticshttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}
