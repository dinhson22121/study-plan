// Package bootstrap assembles the HTTP router by registering every domain module
// in dependency order. It lives in its own package (not app) because it imports
// the modules, while the modules import app — keeping app free of cycles. main
// and the end-to-end tests both build the router through here.
package bootstrap

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/analytics"
	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/auth"
	"github.com/son-ngo/edu-app/internal/content"
	"github.com/son-ngo/edu-app/internal/curriculum"
	"github.com/son-ngo/edu-app/internal/goal"
	"github.com/son-ngo/edu-app/internal/notification"
	"github.com/son-ngo/edu-app/internal/placement"
	"github.com/son-ngo/edu-app/internal/progress"
	"github.com/son-ngo/edu-app/internal/question"
	"github.com/son-ngo/edu-app/internal/quiz"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/internal/studyplan"
	"github.com/son-ngo/edu-app/internal/user"
)

// BuildRouter assembles the Gin engine, global middleware, the health check, and
// every domain module under /api/v1. It returns the notification module so the
// caller can drive its background workers' lifecycle.
//
// Registration order encodes the deps handoffs:
//   - auth first (sets deps.AuthValidate, used by every protected route)
//   - analytics before notification (sets deps.ReengagementSource)
//   - progress & studyplan after notification (use deps.Notifier)
func BuildRouter(deps *app.Deps) (*gin.Engine, *notification.Module) {
	if deps.Cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(middleware.Logger(deps.Log), middleware.Recovery(deps.Log))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	auth.Register(v1, deps)
	user.Register(v1, deps)
	curriculum.Register(v1, deps)
	question.Register(v1, deps)
	content.Register(v1, deps)
	goal.Register(v1, deps)
	placement.Register(v1, deps)
	quiz.Register(v1, deps)
	analytics.Register(v1, deps) // sets deps.ReengagementSource (before notification)
	notifModule := notification.Register(v1, deps)
	progress.Register(v1, deps)  // uses deps.Notifier (after notification)
	studyplan.Register(v1, deps) // uses deps.Notifier (after notification)

	return router, notifModule
}
