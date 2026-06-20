// Package question wires the question bounded context.
package question

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/question/application"
	"github.com/son-ngo/edu-app/internal/question/infrastructure"
	questionhttp "github.com/son-ngo/edu-app/internal/question/interfaces/http"
)

// Register assembles the question module and mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgRepository(deps.DB)
	svc := application.NewService(repo)
	questionhttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}

// NewService exposes the question service for other modules (placement, quiz).
func NewService(deps *app.Deps) *application.Service {
	return application.NewService(infrastructure.NewPgRepository(deps.DB))
}
