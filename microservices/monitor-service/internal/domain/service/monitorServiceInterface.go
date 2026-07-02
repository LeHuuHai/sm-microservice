package service

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type MonitorServiceInterface interface {
	ProcessHeartbeat(ctx context.Context, hb pkgmodel.Heartbeat) error
	ProcessPingResult(ctx context.Context, res pkgmodel.ResponsePing) error
	SyncServerLifecycle(ctx context.Context, event pkgmodel.ServerEvent, action pkgmodel.ServerActionType) error
	GetLiveStatuses(ctx context.Context, from int, to int) ([]model.LiveStatusWithServerInfo, int, int, int, int, error)
}
