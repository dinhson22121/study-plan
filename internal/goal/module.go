package goal

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/goal/application"
	"github.com/son-ngo/edu-app/internal/goal/infrastructure"
	goalhttp "github.com/son-ngo/edu-app/internal/goal/interfaces/http"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	goalhttp.NewHandler(NewService(deps), deps.AuthValidate).Routes(rg)
}

func NewService(deps *app.Deps) *application.Service {
	return application.NewService(infrastructure.NewPgRepository(deps.DB))
}
