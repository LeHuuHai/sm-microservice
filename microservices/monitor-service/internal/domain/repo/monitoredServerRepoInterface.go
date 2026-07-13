package repo

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
)

type MonitoredServerRepositoryInterface interface {
	Create(ctx context.Context, s *model.MonitoredServer) error
	CreateBatch(ctx context.Context, servers []model.MonitoredServer) error
	Update(ctx context.Context, s *model.MonitoredServer) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.MonitoredServer, error)
	List(ctx context.Context) ([]model.MonitoredServer, error)
}
