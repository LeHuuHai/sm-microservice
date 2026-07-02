package postgres

import (
	"context"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"gorm.io/gorm"
)

type LiveStatusRepo struct {
	db *gorm.DB
}

func NewLiveStatusRepository(db *gorm.DB) repo.LiveStatusRepositoryInterface {
	return &LiveStatusRepo{db: db}
}

func (r *LiveStatusRepo) Create(ctx context.Context, s *model.LiveStatus) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *LiveStatusRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("server_id = ?", id).
		Delete(&model.LiveStatus{}).
		Error
}

func (r *LiveStatusRepo) List(ctx context.Context) ([]model.LiveStatus, error) {
	var list []model.LiveStatus
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *LiveStatusRepo) BulkUpdateLiveStatus(ctx context.Context, items []model.LiveStatus) error {
	if len(items) == 0 {
		return nil
	}

	tx := r.db.WithContext(ctx)
	for _, item := range items {
		if err := tx.Model(&model.LiveStatus{}).
			Where("server_id = ?", item.ServerID).
			Updates(map[string]any{
				"status":            item.Status,
				"last_ping_at":      item.LastPingAt,
				"last_heartbeat_at": item.LastHeartbeatAt,
			}).Error; err != nil {
			slog.Warn("Update live status failed", "err", err)
		}
	}

	return nil
}

func (r *LiveStatusRepo) ListWithPagination(ctx context.Context, from int, to int) ([]model.LiveStatusWithServerInfo, int, error) {
	var results []model.LiveStatusWithServerInfo
	var total int64

	query := r.db.WithContext(ctx).
		Table("monitored_servers ms").
		Select("ms.server_id, ms.server_name, ms.ipv4, ls.status").
		Joins("LEFT JOIN live_statuses ls ON ms.server_id = ls.server_id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("ms.server_name ASC").
		Offset(from).
		Limit(to - from).
		Scan(&results).Error

	return results, int(total), err
}

func (r *LiveStatusRepo) GetStatusSummary(ctx context.Context) (online int, offline int, unknown int, err error) {
	type Result struct {
		Status string
		Count  int
	}
	var results []Result
	err = r.db.WithContext(ctx).
		Table("live_statuses").
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error
	
	if err != nil {
		return 0, 0, 0, err
	}

	for _, res := range results {
		switch res.Status {
		case "ONLINE":
			online = res.Count
		case "OFFLINE":
			offline = res.Count
		default:
			unknown += res.Count
		}
	}
	return online, offline, unknown, nil
}
