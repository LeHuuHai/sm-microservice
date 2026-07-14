package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type PingRequestConsumer struct {
	reader *kafka.Reader
}

func NewPingRequestConsumer(r *kafka.Reader) mq.PingRequestConsumerInterface {
	return &PingRequestConsumer{
		reader: r,
	}
}

func (c *PingRequestConsumer) Read(ctx context.Context) (pkgmodel.RequestPing, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return pkgmodel.RequestPing{}, err
	}

	var req pkgmodel.RequestPing
	err = json.Unmarshal(msg.Value, &req)
	if err != nil {
		return pkgmodel.RequestPing{}, err
	}

	return req, nil
}

func (c *PingRequestConsumer) Close() error {
	return c.reader.Close()
}
