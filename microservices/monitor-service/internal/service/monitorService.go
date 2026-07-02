package service

import (
	"context"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type MonitorService struct {
	monitoredServerRepo repo.MonitoredServerRepositoryInterface
	liveStatusRepo      repo.LiveStatusRepositoryInterface
	pgChan              chan model.LiveStatus
	esChan              chan model.StatusLog
}

func NewMonitorService(
	monitoredServerRepo repo.MonitoredServerRepositoryInterface,
	liveStatusRepo repo.LiveStatusRepositoryInterface,
	pgChan chan model.LiveStatus,
	esChan chan model.StatusLog,
) service.MonitorServiceInterface {
	return &MonitorService{
		monitoredServerRepo: monitoredServerRepo,
		liveStatusRepo:      liveStatusRepo,
		pgChan:              pgChan,
		esChan:              esChan,
	}
}

func (s *MonitorService) ProcessHeartbeat(ctx context.Context, hb pkgmodel.Heartbeat) error {
	server := model.LiveStatus{
		ServerID:        hb.ServerID,
		Status:          pkgmodel.StatusOnline,
		LastHeartbeatAt: &hb.Timestamp,
	}
	s.pgChan <- server

	event := model.StatusLog{
		ServerID:  hb.ServerID,
		Status:    pkgmodel.StatusOnline,
		Timestamp: hb.Timestamp,
	}
	s.esChan <- event

	return nil
}

func (s *MonitorService) ProcessPingResult(ctx context.Context, res pkgmodel.ResponsePing) error {
	server := model.LiveStatus{
		ServerID:   res.ServerID,
		Status:     res.Status,
		LastPingAt: res.PingAt,
	}
	s.pgChan <- server

	event := model.StatusLog{
		ServerID:  res.ServerID,
		Status:    res.Status,
		Timestamp: res.PingAt,
	}
	s.esChan <- event

	return nil
}

func (s *MonitorService) SyncServerLifecycle(ctx context.Context, event pkgmodel.ServerEvent, action pkgmodel.ServerActionType) error {
	switch action {
	case pkgmodel.ServerCreateAction:
		slog.Info("Syncing server create event", "server_id", event.ServerID, "server_name", event.ServerName, "ipv4", event.IPv4, "version", event.Version)
		srv := &model.MonitoredServer{
			ServerID:   event.ServerID,
			ServerName: event.ServerName,
			IPv4:       event.IPv4,
			Version:    event.Version,
		}
		if err := s.monitoredServerRepo.Create(ctx, srv); err != nil {
			slog.Error("Failed to create monitored server", "server_id", event.ServerID, "err", err)
			return err
		}

		status := &model.LiveStatus{
			ServerID: event.ServerID,
			Status:   pkgmodel.StatusUnknown,
		}
		return s.liveStatusRepo.Create(ctx, status)

	case pkgmodel.ServerUpdateAction:
		slog.Info("Syncing server update event", "server_id", event.ServerID, "server_name", event.ServerName, "ipv4", event.IPv4, "version", event.Version)
		srv := &model.MonitoredServer{
			ServerID:   event.ServerID,
			ServerName: event.ServerName,
			IPv4:       event.IPv4,
			Version:    event.Version,
		}
		if err := s.monitoredServerRepo.Update(ctx, srv); err != nil {
			slog.Error("Failed to update monitored server", "server_id", event.ServerID, "err", err)
			return err
		}

	case pkgmodel.ServerDeleteAction:
		slog.Info("Syncing server delete event", "server_id", event.ServerID, "server_name", event.ServerName, "ipv4", event.IPv4, "version", event.Version)
		if err := s.monitoredServerRepo.Delete(ctx, event.ServerID); err != nil {
			slog.Error("Failed to delete monitored server", "server_id", event.ServerID, "err", err)
			return err
		}
		return s.liveStatusRepo.Delete(ctx, event.ServerID)
	}
	return nil
}

func (s *MonitorService) GetLiveStatuses(ctx context.Context, from int, to int) ([]model.LiveStatusWithServerInfo, int, int, int, int, error) {
	if to-from <= 0 || from < 0 || to <= 0 {
		return nil, 0, 0, 0, 0, apperr.ErrInvalidPagination
	}
	if to-from > 100 {
		to = from + 100
	}

	results, total, err := s.liveStatusRepo.ListWithPagination(ctx, from, to)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}

	online, offline, unknown, err := s.liveStatusRepo.GetStatusSummary(ctx)
	if err != nil {
		// Log error but don't fail the entire request just because summary failed
		slog.Warn("Failed to get status summary", "err", err)
	}

	return results, total, online, offline, unknown, nil
}
