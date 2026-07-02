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
