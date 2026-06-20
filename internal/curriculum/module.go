// Package curriculum wires the curriculum bounded context.
package curriculum

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/curriculum/application"
	"github.com/son-ngo/edu-app/internal/curriculum/infrastructure"
	curriculumhttp "github.com/son-ngo/edu-app/internal/curriculum/interfaces/http"
)

// Register assembles the curriculum module and mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgRepository(deps.DB)
	svc := application.NewService(repo)
	curriculumhttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}

// NewService exposes the curriculum service for other modules that need catalog
// reads (e.g. studyplan in Phase 4) without importing the HTTP layer.
func NewService(deps *app.Deps) *application.Service {
	return application.NewService(infrastructure.NewPgRepository(deps.DB))
}
