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

func (c *Consumers) Start(ctx context.Context) {
	startConsumer(c, ctx, domain.TopicSchedule, c.schedule.Process)
	startConsumer(c, ctx, domain.TopicSend, c.send.Process)
	startConsumer(c, ctx, domain.TopicResult, c.result.Process)
}

func (c *Consumers) Close() {
	c.wg.Wait()
	for _, r := range c.readers {
		_ = r.Close()
	}
}

func (c *Consumers) newReader(topic string) *kgo.Reader {

	r := c.client.NewReader(kafka.ReaderConfig{Topic: topic, GroupID: c.groupID + "-" + topic})
	c.readers = append(c.readers, r)
	return r
}

func (c *Consumers) deadLetter(ctx context.Context, srcTopic string, key, value []byte) {
	if c.producer == nil {
		return
	}
	if err := c.producer.Publish(ctx, domain.TopicDLQ, key, value); err != nil {
		c.log.Error("kafka dead-letter publish failed", zap.String("src_topic", srcTopic), zap.Error(err))
	}
}

func startConsumer[T any](c *Consumers, ctx context.Context, topic string, handler func(context.Context, T) error) {
	reader := c.newReader(topic)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		runConsumer(ctx, c, reader, handler)
	}()
}

func runConsumer[T any](ctx context.Context, c *Consumers, reader *kgo.Reader, handler func(context.Context, T) error) {
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
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
