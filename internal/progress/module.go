package progress

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/curriculum"
	"github.com/son-ngo/edu-app/internal/progress/application"
	"github.com/son-ngo/edu-app/internal/progress/infrastructure"
	progresshttp "github.com/son-ngo/edu-app/internal/progress/interfaces/http"
	quizdomain "github.com/son-ngo/edu-app/internal/quiz/domain"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	svc := NewService(deps)
	deps.Bus.Subscribe(quizdomain.EventQuizCompleted, svc.HandleQuizCompleted)
	progresshttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}

func NewService(deps *app.Deps) *application.Service {
	repo := infrastructure.NewPgRepository(deps.DB)
	titles := infrastructure.NewTopicTitleAdapter(curriculum.NewService(deps))
	return application.NewService(repo, titles, deps.Notifier, deps.Log)
}
