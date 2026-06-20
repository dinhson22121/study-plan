// Package notification wires the notification bounded context: HTTP routes,
// the Kafka pipeline consumers, the FCM adapter, and the cron scheduler. Because
// it owns background workers, Register returns a *Module whose Start/Stop the
// server lifecycle drives.
package notification

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/notification/application"
	"github.com/son-ngo/edu-app/internal/notification/domain"
	"github.com/son-ngo/edu-app/internal/notification/infrastructure"
	notifhttp "github.com/son-ngo/edu-app/internal/notification/interfaces/http"
	notifkafka "github.com/son-ngo/edu-app/internal/notification/interfaces/kafka"
	"github.com/son-ngo/edu-app/pkg/fcm"
	"go.uber.org/zap"
)

// Module holds the notification background workers so the server can start and
// stop them around the HTTP lifecycle.
type Module struct {
	consumers *notifkafka.Consumers
	scheduler *application.Scheduler
	kafka     interface {
		EnsureTopics(ctx context.Context, partitions int, topics ...string) error
	}
	partitions int
	log        *zap.Logger
}

// Register builds the notification module, mounts its routes, and returns the
// module so the caller can Start/Stop its consumers and scheduler.
func Register(rg *gin.RouterGroup, deps *app.Deps) *Module {
	repo := infrastructure.NewPgRepository(deps.DB)
	idem := infrastructure.NewRedisIdempotencyStore(deps.Redis)

	dispatcher := application.NewDispatcher(repo, idem, deps.Producer, deps.Log)
	manager := application.NewManager(repo, dispatcher)

	// Expose the dispatcher to other modules (e.g. studyplan) as a Notifier.
	deps.Notifier = &notifier{dispatcher: dispatcher}

	sender := buildSender(deps)
	fcmAdapter := infrastructure.NewFCMAdapter(sender, repo, deps.Log)

	schedule := application.NewScheduleProcessor(repo, deps.Producer, deps.Log)
	send := application.NewSendProcessor(fcmAdapter, deps.Producer, deps.Log)
	result := application.NewResultProcessor(repo, deps.Producer, deps.Log)

	consumers := notifkafka.NewConsumers(deps.Kafka, deps.Producer, deps.Cfg.Kafka.GroupID, schedule, send, result, deps.Log)
	scheduler := application.NewScheduler(dispatcher, repo, deps.ReengagementSource, deps.Log, deps.Cfg.Timezone)

	// Seed default preferences when a user registers.
	deps.Bus.Subscribe(EventUserRegistered, newPreferenceSeeder(repo, deps.Log).handle)

	notifhttp.NewHandler(manager, deps.AuthValidate).Routes(rg)

	return &Module{
		consumers:  consumers,
		scheduler:  scheduler,
		kafka:      deps.Kafka,
		partitions: deps.Cfg.Kafka.Partitions,
		log:        deps.Log,
	}
}

// Start provisions Kafka topics, launches the consumers, and starts the cron
// scheduler. Consumers run until ctx is cancelled.
func (m *Module) Start(ctx context.Context) error {
	if err := m.kafka.EnsureTopics(ctx, m.partitions, domain.AllTopics()...); err != nil {
		return err
	}
	m.consumers.Start(ctx)
	return m.scheduler.Start()
}

// Stop halts the scheduler and closes the consumers.
func (m *Module) Stop() {
	m.scheduler.Stop()
	m.consumers.Close()
}

// buildSender returns a real FCM sender when credentials initialize, otherwise a
// logging fallback so local dev runs without Firebase.
func buildSender(deps *app.Deps) interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
	IsTokenInvalid(err error) bool
} {
	client, err := fcm.New(context.Background(), fcm.Config{
		CredentialsFile: deps.Cfg.FCM.CredentialsFile,
		ProjectID:       deps.Cfg.FCM.ProjectID,
	})
	if err != nil {
		deps.Log.Warn("FCM not configured, using log fallback sender", zap.Error(err))
		return infrastructure.NewLogSender(deps.Log)
	}
	return client
}
