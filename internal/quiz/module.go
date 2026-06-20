// Package quiz wires the quiz bounded context. It reads the question bank to
// assemble and grade quizzes and emits quiz.completed for progress/analytics.
package quiz

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/question"
	"github.com/son-ngo/edu-app/internal/quiz/application"
	"github.com/son-ngo/edu-app/internal/quiz/infrastructure"
	quizhttp "github.com/son-ngo/edu-app/internal/quiz/interfaces/http"
)

// Register assembles the quiz module and mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	quizhttp.NewHandler(NewService(deps), deps.AuthValidate).Routes(rg)
}

// NewService builds the quiz service. Exposed so analytics can read quiz results.
func NewService(deps *app.Deps) *application.Service {
	repo := infrastructure.NewPgRepository(deps.DB)
	source := infrastructure.NewQuestionSourceAdapter(question.NewService(deps))
	return application.NewService(repo, source, deps.Bus)
}
