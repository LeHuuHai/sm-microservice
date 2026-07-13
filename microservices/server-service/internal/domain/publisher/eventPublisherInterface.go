package publisher

import (
	"context"
)

type EventPublisherInterface interface {
	PublishEvent(ctx context.Context, topic string, payload []byte) error
}
