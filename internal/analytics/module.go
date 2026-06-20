// Package analytics wires the analytics bounded context. It reads progress and
// quiz, records activity from quiz.completed, and exposes the inactive-user feed
// to the notification re-engagement scheduler. Register before notification.
package analytics

import (
	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/analytics/application"
	"github.com/son-ngo/edu-app/internal/analytics/infrastructure"
	analyticshttp "github.com/son-ngo/edu-app/internal/analytics/interfaces/http"
	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/progress"
	"github.com/son-ngo/edu-app/internal/quiz"
	quizdomain "github.com/son-ngo/edu-app/internal/quiz/domain"
)

// Register assembles the analytics module, subscribes it to quiz.completed for
// activity tracking, publishes the re-engagement source, and mounts its routes.
func Register(rg *gin.RouterGroup, deps *app.Deps) {
	activityRepo := infrastructure.NewPgActivityRepo(deps.DB)
	progressReader := infrastructure.NewProgressReaderAdapter(progress.NewService(deps))
	quizReader := infrastructure.NewQuizReaderAdapter(quiz.NewService(deps))

	svc := application.NewService(activityRepo, progressReader, quizReader, deps.Log)

	deps.Bus.Subscribe(quizdomain.EventQuizCompleted, svc.HandleQuizCompleted)
	deps.ReengagementSource = svc // consumed by notification's re-engagement cron

	analyticshttp.NewHandler(svc, deps.AuthValidate).Routes(rg)
}
