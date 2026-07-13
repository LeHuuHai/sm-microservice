package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/publisher"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/repo"
)

type OutboxPoller struct {
	outboxRepo repo.OutboxRepositoryInterface
	publisher  publisher.EventPublisherInterface
}

func NewOutboxPoller(r repo.OutboxRepositoryInterface, p publisher.EventPublisherInterface) *OutboxPoller {
	return &OutboxPoller{
		outboxRepo: r,
		publisher:  p,
	}
}

func (p *OutboxPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Outbox poller stopped")
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *OutboxPoller) processBatch(ctx context.Context) {
	events, err := p.outboxRepo.FetchPendingEvents(ctx, 100)
	if err != nil {
		slog.Error("Failed to fetch outbox events", "error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	var successfulIDs []string

	for _, ev := range events {
		err := p.publisher.PublishEvent(ctx, string(ev.Topic), ev.Payload)

		if err != nil {
			slog.Error("Failed to publish outbox event", "error", err, "id", ev.ID)
			continue
		}

		successfulIDs = append(successfulIDs, ev.ID)
	}

	if len(successfulIDs) > 0 {
		if err := p.outboxRepo.MarkEventsDone(ctx, successfulIDs); err != nil {
			slog.Error("Failed to mark outbox events as done", "error", err)
		}
	}
}
