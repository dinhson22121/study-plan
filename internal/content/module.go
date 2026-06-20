// Package content wires the content bounded context.
package content

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/content/application"
	"github.com/son-ngo/edu-app/internal/content/infrastructure"
	contenthttp "github.com/son-ngo/edu-app/internal/content/interfaces/http"
)

// Register assembles the content module and mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgRepository(deps.DB)
	svc := application.NewService(repo)
	contenthttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}
