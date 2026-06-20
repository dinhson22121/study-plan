// Command server is the edu-app modular-monolith entrypoint. It loads config,
// connects infrastructure, wires every domain module through a shared Deps
// container, and serves HTTP with graceful shutdown.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/config"
	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/auth"
	"github.com/son-ngo/edu-app/internal/notification"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/internal/user"
	"github.com/son-ngo/edu-app/pkg/kafka"
	"github.com/son-ngo/edu-app/pkg/postgres"
	"github.com/son-ngo/edu-app/pkg/redis"
)

func main() {
	log, _ := zap.NewProduction()
	defer func() { _ = log.Sync() }()

	if err := run(log); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}

func run(log *zap.Logger) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := postgres.Connect(ctx, postgres.Config{
		URL: cfg.Postgres.URL, MaxConns: cfg.Postgres.MaxConns, MinConns: cfg.Postgres.MinConns,
	})
	if err != nil {
		return err
	}
	defer db.Close()

	rdb, err := redis.Connect(ctx, cfg.Redis.URL)
	if err != nil {
		return err
	}
	defer func() { _ = rdb.Close() }()

	kafkaClient := kafka.NewClient(cfg.Kafka.Brokers)
	producer := kafkaClient.NewProducer()
	defer func() { _ = producer.Close() }()

	deps := &app.Deps{
		Cfg:      cfg,
		DB:       db,
		Redis:    rdb,
		Kafka:    kafkaClient,
		Producer: producer,
		Bus:      eventbus.New(),
		Log:      log,
	}

	// Background workers (notification consumers + scheduler) run until cancelled.
	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	router, notifModule := buildRouter(deps)
	if err := notifModule.Start(workerCtx); err != nil {
		return err
	}
	defer notifModule.Stop()

	srv := &http.Server{Addr: cfg.Port, Handler: router}
	return serveWithGracefulShutdown(srv, log)
}

// buildRouter assembles the Gin engine, global middleware, the health check, and
// every domain module under /api/v1. It returns the notification module so the
// caller can drive its background workers' lifecycle.
func buildRouter(deps *app.Deps) (*gin.Engine, *notification.Module) {
	if deps.Cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(middleware.Logger(deps.Log), middleware.Recovery(deps.Log))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	// Register all domain modules. New modules are added here as phases land.
	auth.Register(v1, deps)
	user.Register(v1, deps)
	notifModule := notification.Register(v1, deps)

	return router, notifModule
}

func serveWithGracefulShutdown(srv *http.Server, log *zap.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("http server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-stop:
		log.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
