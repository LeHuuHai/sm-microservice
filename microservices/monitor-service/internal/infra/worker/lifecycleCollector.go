package worker

import (
	"context"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type LifecycleConsumer struct {
	consumer mq.LifecycleConsumerInterface
	svc      service.MonitorServiceInterface
}

func NewLifecycleConsumer(consumer mq.LifecycleConsumerInterface, svc service.MonitorServiceInterface) *LifecycleConsumer {
	return &LifecycleConsumer{
		consumer: consumer,
		svc:      svc,
	}
}

func (c *LifecycleConsumer) Start(ctx context.Context) {
	for {
		events, action, commitFunc, err := c.consumer.Read(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				slog.Error("LifecycleConsumer failed to read message", "err", err)
				continue
			}
		}
		slog.Info("Received server lifecycle events", "action", action, "count", len(events))

		var syncErr error
		if action == pkgmodel.ServerBatchCreateAction {
			syncErr = c.svc.SyncServerLifecycleBatch(ctx, events)
		} else {
			err = c.svc.SyncServerLifecycle(ctx, events[0], action)
			if err != nil {
				slog.Error("Failed to sync server lifecycle event", "action", action, "server_id", events[0].ServerID, "err", err)
				syncErr = err
			}
		}

		if syncErr != nil {
			slog.Error("Sync failed, skipping commit to retry", "action", action, "err", syncErr)
			continue
		}

		if err := commitFunc(ctx); err != nil {
			slog.Error("Failed to commit lifecycle message", "err", err)
		} else {
			slog.Info("Successfully synced server lifecycle event", "action", action, "count", len(events))
		}
	}
}
