package bootstrap

import (
	"context"
	"net/http"
	"time"

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
	"github.com/son-ngo/edu-app/internal/shared/metrics"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/internal/studyplan"
	"github.com/son-ngo/edu-app/internal/user"
)

func BuildRouter(deps *app.Deps) (*gin.Engine, *notification.Module) {
	if deps.Cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	if deps.Metrics == nil {
		deps.Metrics = metrics.New()
	}

	router := gin.New()
	router.Use(middleware.Logger(deps.Log), middleware.Recovery(deps.Log), deps.Metrics.Middleware())

	router.GET("/metrics", gin.WrapH(deps.Metrics.Handler()))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		checks := gin.H{"postgres": "ok", "redis": "ok"}
		ready := true
		if err := deps.DB.Ping(ctx); err != nil {
			checks["postgres"] = "down"
			ready = false
		}
		if err := deps.Redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "down"
			ready = false
		}
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"ready": ready, "checks": checks})
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
	analytics.Register(v1, deps)
	notifModule := notification.Register(v1, deps)
	progress.Register(v1, deps)
	studyplan.Register(v1, deps)

	return router, notifModule
}
