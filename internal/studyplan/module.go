package studyplan

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/curriculum"
	"github.com/son-ngo/edu-app/internal/goal"
	"github.com/son-ngo/edu-app/internal/placement"
	"github.com/son-ngo/edu-app/internal/studyplan/application"
	"github.com/son-ngo/edu-app/internal/studyplan/infrastructure"
	studyplanhttp "github.com/son-ngo/edu-app/internal/studyplan/interfaces/http"
)

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgRepository(deps.DB)
	topics := infrastructure.NewTopicSourceAdapter(curriculum.NewService(deps))
	levels := infrastructure.NewLevelSourceAdapter(placement.NewService(deps))
	goals := infrastructure.NewGoalSourceAdapter(goal.NewService(deps))
	reminder := infrastructure.NewReminderAdapter(deps.Notifier)

	svc := application.NewService(repo, topics, levels, goals, reminder)
	studyplanhttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}
