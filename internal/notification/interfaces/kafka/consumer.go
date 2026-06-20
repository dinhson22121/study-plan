// Package notifkafka wires Kafka topic consumers to the notification pipeline
// processors. Each topic runs its own consumer goroutine.
package notifkafka

import (
	"context"
	"encoding/json"

	kgo "github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/application"
	"github.com/son-ngo/edu-app/internal/notification/domain"
	"github.com/son-ngo/edu-app/pkg/kafka"
)

// Consumers owns the consumer goroutines for the three pipeline stages.
type Consumers struct {
	client   *kafka.Client
	groupID  string
	schedule *application.ScheduleProcessor
	send     *application.SendProcessor
	result   *application.ResultProcessor
	log      *zap.Logger
	readers  []*kgo.Reader
}

// NewConsumers builds the consumer set.
func NewConsumers(
	client *kafka.Client,
	groupID string,
	schedule *application.ScheduleProcessor,
	send *application.SendProcessor,
	result *application.ResultProcessor,
	log *zap.Logger,
) *Consumers {
	return &Consumers{client: client, groupID: groupID, schedule: schedule, send: send, result: result, log: log}
}

// Start launches one consumer goroutine per pipeline topic. They run until ctx
// is cancelled. Call Close afterwards to release readers.
func (c *Consumers) Start(ctx context.Context) {
	go runConsumer(ctx, c.newReader(domain.TopicSchedule), c.log, func(ctx context.Context, m domain.ScheduleMessage) error {
		return c.schedule.Process(ctx, m)
	})
	go runConsumer(ctx, c.newReader(domain.TopicSend), c.log, func(ctx context.Context, m domain.SendMessage) error {
		return c.send.Process(ctx, m)
	})
	go runConsumer(ctx, c.newReader(domain.TopicResult), c.log, func(ctx context.Context, m domain.ResultMessage) error {
		return c.result.Process(ctx, m)
	})
}

// Close releases all readers.
func (c *Consumers) Close() {
	for _, r := range c.readers {
		_ = r.Close()
	}
}

func (c *Consumers) newReader(topic string) *kgo.Reader {
	// A per-topic group id keeps offsets independent across stages.
	r := c.client.NewReader(kafka.ReaderConfig{Topic: topic, GroupID: c.groupID + "-" + topic})
	c.readers = append(c.readers, r)
	return r
}

// runConsumer reads messages, decodes them into T, and invokes handler. It
// commits after each message (success or handled error) so a poison message
// never blocks the partition; processors record failures to the log/DLQ.
func runConsumer[T any](ctx context.Context, reader *kgo.Reader, log *zap.Logger, handler func(context.Context, T) error) {
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // shutting down
			}
			log.Error("kafka fetch failed", zap.String("topic", reader.Config().Topic), zap.Error(err))
			continue
		}

		var payload T
		if err := json.Unmarshal(m.Value, &payload); err != nil {
			log.Error("kafka message decode failed", zap.String("topic", m.Topic), zap.Error(err))
		} else if herr := handler(ctx, payload); herr != nil {
			log.Error("kafka message handler failed", zap.String("topic", m.Topic), zap.Error(herr))
		}

		if err := reader.CommitMessages(ctx, m); err != nil && ctx.Err() == nil {
			log.Error("kafka commit failed", zap.String("topic", m.Topic), zap.Error(err))
		}
	}
}
