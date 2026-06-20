package placement

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/curriculum"
	"github.com/son-ngo/edu-app/internal/placement/application"
	"github.com/son-ngo/edu-app/internal/placement/infrastructure"
	placementhttp "github.com/son-ngo/edu-app/internal/placement/interfaces/http"
	"github.com/son-ngo/edu-app/internal/question"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	placementhttp.NewHandler(NewService(deps), deps.AuthValidate).Routes(rg)
}

func NewService(deps *app.Deps) *application.Service {
	repo := infrastructure.NewPgRepository(deps.DB)
	source := infrastructure.NewQuestionSourceAdapter(curriculum.NewService(deps), question.NewService(deps))
	return application.NewService(repo, source, deps.Bus)
}
