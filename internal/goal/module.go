// Package goal wires the goal bounded context.
package goal

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/goal/application"
	"github.com/son-ngo/edu-app/internal/goal/infrastructure"
	goalhttp "github.com/son-ngo/edu-app/internal/goal/interfaces/http"
)

// Register assembles the goal module and mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	goalhttp.NewHandler(NewService(deps), deps.AuthValidate).Routes(rg)
}

// NewService exposes the goal service for other modules (studyplan).
func NewService(deps *app.Deps) *application.Service {
	return application.NewService(infrastructure.NewPgRepository(deps.DB))
}
