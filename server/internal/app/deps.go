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

type Notifier interface {
	EnqueueReminder(ctx context.Context, userID, notifType, templateCode string, vars map[string]string, idempotencyKey string) error
}

type ReengagementSource interface {
	InactiveUserIDs(ctx context.Context, days int) ([]string, error)
}

type Deps struct {
	Cfg      *config.Config
	DB       *pgxpool.Pool
	Redis    *goredis.Client
	Kafka    *kafka.Client
	Producer kafka.Producer
	Bus      *eventbus.Bus
	Log      *zap.Logger

	AuthValidate middleware.TokenValidator

	AuthRateLimiter middleware.RateLimiter

	Notifier Notifier

	ReengagementSource ReengagementSource
}
