// Package app holds the shared dependency container passed to every domain
// module at registration, plus the server assembly. Modules depend on this
// package for the Deps type; this package never imports modules, so no cycle.
package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"context"

	"github.com/son-ngo/edu-app/config"
	"github.com/son-ngo/edu-app/internal/shared/eventbus"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
	"github.com/son-ngo/edu-app/pkg/kafka"
)

// Notifier enqueues a notification through the notification pipeline (preference
// gate + idempotency + Kafka). The notification module sets deps.Notifier on
// Register; upstream modules (e.g. studyplan) use it to trigger reminders
// without importing notification internals.
type Notifier interface {
	EnqueueReminder(ctx context.Context, userID, notifType, templateCode string, vars map[string]string, idempotencyKey string) error
}

// ReengagementSource lists users who have been inactive for at least the given
// number of days. The analytics module sets deps.ReengagementSource on Register;
// the notification re-engagement scheduler consumes it.
type ReengagementSource interface {
	InactiveUserIDs(ctx context.Context, days int) ([]string, error)
}

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

	// Notifier enqueues notifications. The notification module populates this on
	// Register; modules registered afterward use it to trigger reminders.
	Notifier Notifier

	// ReengagementSource lists inactive users. The analytics module populates
	// this on Register; the notification re-engagement scheduler consumes it.
	// analytics must register before notification.
	ReengagementSource ReengagementSource
}
