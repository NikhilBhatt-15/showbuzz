package platform

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Kafka topics shared between producers and the notify consumer.
const (
	TopicBookingConfirmed = "booking.confirmed"
	TopicBookingCancelled = "booking.cancelled"
	TopicOTPRequested     = "otp.requested"
	TopicClickstream      = "clickstream"
)

func brokers() []string { return strings.Split(Env("KAFKA_BROKERS", "localhost:9092"), ",") }

// Producer publishes JSON events asynchronously.
type Producer struct{ cl *kgo.Client }

func NewProducer() (*Producer, error) {
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers()...), kgo.AllowAutoTopicCreation())
	if err != nil {
		return nil, err
	}
	return &Producer{cl: cl}, nil
}

// Publish enqueues v on topic. Delivery failures are logged, not returned, so
// a broker outage never fails a user-facing request.
func (p *Producer) Publish(ctx context.Context, topic, key string, v any) {
	if p == nil {
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("marshal event", "topic", topic, "err", err)
		return
	}
	p.cl.Produce(context.WithoutCancel(ctx), &kgo.Record{Topic: topic, Key: []byte(key), Value: b},
		func(_ *kgo.Record, err error) {
			if err != nil {
				slog.Error("publish event", "topic", topic, "err", err)
			}
		})
}

func (p *Producer) Close() {
	if p != nil {
		p.cl.Close()
	}
}

// NewConsumer joins a consumer group on the given topics.
func NewConsumer(group string, topics ...string) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(brokers()...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		kgo.AllowAutoTopicCreation(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
}
