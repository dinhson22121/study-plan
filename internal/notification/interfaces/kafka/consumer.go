// Package notifkafka wires Kafka topic consumers to the notification pipeline
// processors. Each topic runs its own consumer goroutine.
package notifkafka

import (
	"context"
	"encoding/json"
	"sync"

	kgo "github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/notification/application"
	"github.com/son-ngo/edu-app/internal/notification/domain"
	"github.com/son-ngo/edu-app/pkg/kafka"
)

// Consumers owns the consumer goroutines for the three pipeline stages.
type Consumers struct {
	client   *kafka.Client
	producer kafka.Producer
	groupID  string
	schedule *application.ScheduleProcessor
	send     *application.SendProcessor
	result   *application.ResultProcessor
	log      *zap.Logger
	readers  []*kgo.Reader
	wg       sync.WaitGroup
}

// NewConsumers builds the consumer set. The producer is used to dead-letter
// messages that cannot be decoded.
func NewConsumers(
	client *kafka.Client,
	producer kafka.Producer,
	groupID string,
	schedule *application.ScheduleProcessor,
	send *application.SendProcessor,
	result *application.ResultProcessor,
	log *zap.Logger,
) *Consumers {
	return &Consumers{client: client, producer: producer, groupID: groupID, schedule: schedule, send: send, result: result, log: log}
}

// Start launches one consumer goroutine per pipeline topic. They run until ctx
// is cancelled. Call Close afterwards to wait for the goroutines and release
// readers.
func (c *Consumers) Start(ctx context.Context) {
	startConsumer(c, ctx, domain.TopicSchedule, c.schedule.Process)
	startConsumer(c, ctx, domain.TopicSend, c.send.Process)
	startConsumer(c, ctx, domain.TopicResult, c.result.Process)
}

// Close waits for the consumer goroutines to finish, then releases readers.
func (c *Consumers) Close() {
	c.wg.Wait()
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

// deadLetter forwards an undecodable message to the DLQ for manual review.
func (c *Consumers) deadLetter(ctx context.Context, srcTopic string, key, value []byte) {
	if c.producer == nil {
		return
	}
	if err := c.producer.Publish(ctx, domain.TopicDLQ, key, value); err != nil {
		c.log.Error("kafka dead-letter publish failed", zap.String("src_topic", srcTopic), zap.Error(err))
	}
}

// startConsumer spawns a tracked consumer goroutine for one topic. It is a free
// function because Go methods cannot have type parameters.
func startConsumer[T any](c *Consumers, ctx context.Context, topic string, handler func(context.Context, T) error) {
	reader := c.newReader(topic)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		runConsumer(ctx, c, reader, handler)
	}()
}

// runConsumer reads messages, decodes them into T, and invokes handler. A
// message that cannot be decoded is dead-lettered (not silently dropped) and a
// handler error is logged; in both cases the offset is committed so a poison
// message never blocks the partition.
func runConsumer[T any](ctx context.Context, c *Consumers, reader *kgo.Reader, handler func(context.Context, T) error) {
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // shutting down
			}
			c.log.Error("kafka fetch failed", zap.String("topic", reader.Config().Topic), zap.Error(err))
			continue
		}

		var payload T
		if derr := json.Unmarshal(m.Value, &payload); derr != nil {
			c.log.Error("kafka message decode failed; dead-lettering",
				zap.String("topic", m.Topic), zap.Error(derr))
			c.deadLetter(ctx, m.Topic, m.Key, m.Value)
		} else if herr := handler(ctx, payload); herr != nil {
			c.log.Error("kafka message handler failed", zap.String("topic", m.Topic), zap.Error(herr))
		}

		if err := reader.CommitMessages(ctx, m); err != nil && ctx.Err() == nil {
			c.log.Error("kafka commit failed", zap.String("topic", m.Topic), zap.Error(err))
		}
	}
}
