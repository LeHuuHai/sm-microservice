package kafka

import (
	"context"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/publisher"
	"github.com/segmentio/kafka-go"
)

type serverEventPublisher struct {
	writer *kafka.Writer
}

func NewServerEventPublisher(w *kafka.Writer) publisher.EventPublisherInterface {
	return &serverEventPublisher{
		writer: w,
	}
}

func (p *serverEventPublisher) PublishEvent(ctx context.Context, topic string, payload []byte) error {
	msg := kafka.Message{
		Value:   payload,
		Headers: []kafka.Header{{Key: "event_type", Value: []byte(topic)}},
	}
	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		slog.Error("Failed to publish event to Kafka", "error", err)
	}
	slog.Info("Publishing event to Kafka", "event_type", topic)
	return err
}
