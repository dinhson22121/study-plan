// Package progress wires the progress bounded context. It subscribes to
// quiz.completed to update mastery/streaks/achievements and pushes achievement
// notifications. Must be registered after notification (needs deps.Notifier).
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

// Register assembles the progress module, subscribes it to quiz.completed, and
// mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	svc := NewService(deps)
	deps.Bus.Subscribe(quizdomain.EventQuizCompleted, svc.HandleQuizCompleted)
	progresshttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}

// NewService builds the progress service. Exposed so analytics can read progress.
func NewService(deps *app.Deps) *application.Service {
	repo := infrastructure.NewPgRepository(deps.DB)
	titles := infrastructure.NewTopicTitleAdapter(curriculum.NewService(deps))
	return application.NewService(repo, titles, deps.Notifier, deps.Log)
}
