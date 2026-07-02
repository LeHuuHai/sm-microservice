package repo

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
)

type LiveStatusRepositoryInterface interface {
	Create(ctx context.Context, s *model.LiveStatus) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]model.LiveStatus, error)
	BulkUpdateLiveStatus(ctx context.Context, items []model.LiveStatus) error
	ListWithPagination(ctx context.Context, from int, to int) ([]model.LiveStatusWithServerInfo, int, error)
	GetStatusSummary(ctx context.Context) (online int, offline int, unknown int, err error)
}
