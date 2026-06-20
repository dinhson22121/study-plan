// Package placement wires the placement bounded context. It composes the
// curriculum and question services (read-only) to assemble and grade tests.
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

// Register assembles the placement module and mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	placementhttp.NewHandler(NewService(deps), deps.AuthValidate).Routes(rg)
}

// NewService builds the placement service, wiring the question-bank source from
// the curriculum and question services. Exposed so studyplan can read levels.
func NewService(deps *app.Deps) *application.Service {
	repo := infrastructure.NewPgRepository(deps.DB)
	source := infrastructure.NewQuestionSourceAdapter(curriculum.NewService(deps), question.NewService(deps))
	return application.NewService(repo, source, deps.Bus)
}
