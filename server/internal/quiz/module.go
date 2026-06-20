package quiz

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/question"
	"github.com/son-ngo/edu-app/internal/quiz/application"
	"github.com/son-ngo/edu-app/internal/quiz/infrastructure"
	quizhttp "github.com/son-ngo/edu-app/internal/quiz/interfaces/http"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	quizhttp.NewHandler(NewService(deps), deps.AuthValidate).Routes(rg)
}

func NewService(deps *app.Deps) *application.Service {
	repo := infrastructure.NewPgRepository(deps.DB)
	source := infrastructure.NewQuestionSourceAdapter(question.NewService(deps))
	return application.NewService(repo, source, deps.Bus)
}
