// Package app holds the shared dependency container passed to every domain
// module at registration, plus the server assembly. Modules depend on this
// package for the Deps type; this package never imports modules, so no cycle.
package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/config"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/pkg/kafka"
)

// Deps is the set of infrastructure dependencies every module may need. A module
// uses only the fields relevant to it. This keeps each module's Register
// signature uniform and the wiring in main.go declarative.
type Deps struct {
	Cfg      *config.Config
	DB       *pgxpool.Pool
	Redis    *goredis.Client
	Kafka    *kafka.Client
	Producer kafka.Producer
	Bus      *eventbus.Bus
	Log      *zap.Logger

	// AuthValidate validates access tokens. The auth module populates this on
	// Register; modules registered afterward use it to guard protected routes.
	AuthValidate middleware.TokenValidator
}
