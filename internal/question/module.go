package question

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/question/application"
	"github.com/son-ngo/edu-app/internal/question/infrastructure"
	questionhttp "github.com/son-ngo/edu-app/internal/question/interfaces/http"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgRepository(deps.DB)
	svc := application.NewService(repo)
	questionhttp.NewHandler(svc, deps.AuthValidate).Routes(rg)

	draftSvc := application.NewDraftService(infrastructure.NewPgDraftRepository(deps.DB), svc)
	questionhttp.NewAdminDraftHandler(draftSvc, deps.AuthValidate).Routes(rg)
}

func NewService(deps *app.Deps) *application.Service {
	return application.NewService(infrastructure.NewPgRepository(deps.DB))
}
