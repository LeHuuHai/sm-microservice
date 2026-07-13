package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type LifecycleConsumer struct {
	reader *kafka.Reader
}

func NewLifecycleConsumer(r *kafka.Reader) mq.LifecycleConsumerInterface {
	return &LifecycleConsumer{
		reader: r,
	}
}

func (c *LifecycleConsumer) Read(ctx context.Context) ([]pkgmodel.ServerEvent, pkgmodel.ServerActionType, func(context.Context) error, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	eventType := ""
	for _, h := range msg.Headers {
		if h.Key == "event_type" {
			eventType = string(h.Value)
			break
		}
	}

	var action pkgmodel.ServerActionType
	var events []pkgmodel.ServerEvent
	switch pkgmodel.ServerEventType(eventType) {
	case pkgmodel.ServerCreateEvent:
		action = pkgmodel.ServerCreateAction
		var event pkgmodel.ServerEvent
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			return nil, "", nil, err
		}
		events = []pkgmodel.ServerEvent{event}
	case pkgmodel.ServerUpdateEvent:
		action = pkgmodel.ServerUpdateAction
		var event pkgmodel.ServerEvent
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			return nil, "", nil, err
		}
		events = []pkgmodel.ServerEvent{event}
	case pkgmodel.ServerDeleteEvent:
		action = pkgmodel.ServerDeleteAction
		var event pkgmodel.ServerEvent
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			return nil, "", nil, err
		}
		events = []pkgmodel.ServerEvent{event}
	case pkgmodel.ServerBatchCreateEvent:
		action = pkgmodel.ServerBatchCreateAction
		err = json.Unmarshal(msg.Value, &events)
		if err != nil {
			return nil, "", nil, err
		}
	}

	return events, pkgmodel.ServerActionType(action), func(ctx context.Context) error {
		return c.reader.CommitMessages(ctx, msg)
	}, nil
}

func (c *LifecycleConsumer) Close() error {
	return c.reader.Close()
}
