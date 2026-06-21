package user

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	authdomain "github.com/son-ngo/edu-app/internal/auth/domain"
	"github.com/son-ngo/edu-app/internal/user/application"
	"github.com/son-ngo/edu-app/internal/user/infrastructure"
	userhttp "github.com/son-ngo/edu-app/internal/user/interfaces/http"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgUserRepo(deps.DB)
	svc := application.NewService(repo)

	deps.Bus.Subscribe(authdomain.EventUserRegistered, svc.HandleUserRegistered)

	userhttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}
