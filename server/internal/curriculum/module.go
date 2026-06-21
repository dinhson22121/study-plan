package curriculum

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/curriculum/application"
	"github.com/son-ngo/edu-app/internal/curriculum/infrastructure"
	curriculumhttp "github.com/son-ngo/edu-app/internal/curriculum/interfaces/http"
	"github.com/son-ngo/edu-app/internal/shared/cache"
)

const cacheTTL = 10 * time.Minute

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	curriculumhttp.NewHandler(NewService(deps), deps.AuthValidate).Routes(rg)
}

func NewService(deps *app.Deps) *application.Service {
	return application.NewService(
		infrastructure.NewPgRepository(deps.DB),
		application.WithCache(cache.NewRedisCache(deps.Redis), cacheTTL),
	)
}
