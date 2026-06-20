// Package user wires the user bounded context: it builds the profile service,
// subscribes it to the auth UserRegisteredEvent, and mounts protected routes.
package user

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/user/application"
	"github.com/son-ngo/edu-app/internal/user/infrastructure"
	userhttp "github.com/son-ngo/edu-app/internal/user/interfaces/http"
)

// Register assembles the user module: it subscribes to the registration event so
// a profile is created on signup, and mounts the profile API. Requires
// deps.AuthValidate to be set (auth must register first).
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgUserRepo(deps.DB)
	svc := application.NewService(repo)

	deps.Bus.Subscribe(authdomain.EventUserRegistered, svc.HandleUserRegistered)

	userhttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}
