// Package kafka wraps segmentio/kafka-go with the small surface this app needs:
// a producer, a consumer-reader factory, and topic provisioning. Message
// schemas and topic names live with the modules that own them.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	kgo "github.com/segmentio/kafka-go"
)

// Client is the shared connection configuration (broker list) used to build
// producers and consumers.
type Client struct {
	brokers []string
}

// NewClient returns a Client bound to the given broker addresses.
func NewClient(brokers []string) *Client {
	return &Client{brokers: brokers}
}

// Brokers exposes the configured broker list.
func (c *Client) Brokers() []string { return c.brokers }

// Producer publishes messages to Kafka topics. It is safe for concurrent use.
type Producer interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

type writerProducer struct {
	w *kgo.Writer
}

// NewProducer builds a Producer. A single Writer with Topic unset lets callers
// publish to any topic per-message, which suits a multi-topic app.
func (c *Client) NewProducer() Producer {
	return &writerProducer{
		w: &kgo.Writer{
			Addr:                   kgo.TCP(c.brokers...),
			Balancer:               &kgo.Hash{}, // key-based partitioning for ordering
			RequiredAcks:           kgo.RequireAll,
			AllowAutoTopicCreation: false,
			BatchTimeout:           10 * time.Millisecond,
		},
	}
}

func (p *writerProducer) Publish(ctx context.Context, topic string, key, value []byte) error {
	err := p.w.WriteMessages(ctx, kgo.Message{Topic: topic, Key: key, Value: value})
	if err != nil {
		return fmt.Errorf("kafka publish to %s: %w", topic, err)
	}
	return nil
}

func (p *writerProducer) Close() error { return p.w.Close() }

// ReaderConfig configures a consumer Reader.
type ReaderConfig struct {
	Topic   string
	GroupID string
}

// NewReader builds a consumer Reader for a topic+group. The caller owns the read
// loop and must Close the reader when done.
func (c *Client) NewReader(cfg ReaderConfig) *kgo.Reader {
	return kgo.NewReader(kgo.ReaderConfig{
		Brokers:        c.brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // commit synchronously via CommitMessages for at-least-once
	})
}

// EnsureTopics creates the given topics if they do not already exist, using the
// supplied partition count and replication factor 1 (single-broker dev). It is
// idempotent: topics that already exist are left untouched.
func (c *Client) EnsureTopics(ctx context.Context, partitions int, topics ...string) error {
	if len(c.brokers) == 0 {
		return errors.New("no kafka brokers configured")
	}
	conn, err := kgo.DialContext(ctx, "tcp", c.brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get controller: %w", err)
	}
	ctrlConn, err := kgo.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("dial controller: %w", err)
	}
	defer ctrlConn.Close()

	cfgs := make([]kgo.TopicConfig, 0, len(topics))
	for _, t := range topics {
		cfgs = append(cfgs, kgo.TopicConfig{
			Topic:             t,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		})
	}
	if err := ctrlConn.CreateTopics(cfgs...); err != nil {
		return fmt.Errorf("create topics: %w", err)
	}
	return nil
}
