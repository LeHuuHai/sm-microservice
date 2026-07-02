package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/api"
	domainservice "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
)

type MonitorRestHandler struct {
	reportSvc  domainservice.ReportServiceInterface
	monitorSvc domainservice.MonitorServiceInterface
}

func NewMonitorRestHandler(reportSvc domainservice.ReportServiceInterface, monitorSvc domainservice.MonitorServiceInterface) *MonitorRestHandler {
	return &MonitorRestHandler{
		reportSvc:  reportSvc,
		monitorSvc: monitorSvc,
	}
}

// Generate server report
// (POST /monitor/report)
func (handler *MonitorRestHandler) GenerateServerReport(ctx context.Context, request api.GenerateServerReportRequestObject) (api.GenerateServerReportResponseObject, error) {
	receivers := make([]string, len(request.Body.Receivers))
	for i, r := range request.Body.Receivers {
		receivers[i] = string(r)
	}

	req := model.GenServerReportRequest{
		From:      request.Body.From,
		To:        request.Body.To,
		Receivers: receivers,
	}

	slog.Info("handler: generate server report",
		slog.Any("from", req.From),
		slog.Any("to", req.To),
		slog.Any("receivers", req.Receivers),
	)

	err := handler.reportSvc.ReportServer(ctx, req)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidTimeRange) || errors.Is(err, apperr.ErrInvalidEmail) {
			slog.Warn("invalid request", slog.Any("err", err))
			return api.GenerateServerReport400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		slog.Error("failed to generate server report", slog.Any("err", err))
		return api.GenerateServerReport500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	slog.Info("handler: generate server report accepted",
		slog.Any("receivers", req.Receivers),
	)

	return api.GenerateServerReport202Response{}, nil
}

// Get paginated live status of servers
// (GET /monitor/livestatus)
func (handler *MonitorRestHandler) GetMonitorLivestatus(ctx context.Context, request api.GetMonitorLivestatusRequestObject) (api.GetMonitorLivestatusResponseObject, error) {
	params := request.Params

	results, total, totalOnline, totalOffline, totalUnknown, err := handler.monitorSvc.GetLiveStatuses(ctx, int(params.From), int(params.To))
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidPagination) {
			slog.Warn("invalid pagination", slog.Any("err", err))
			return api.GetMonitorLivestatus400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		slog.Error("failed to get live status list", slog.Any("err", err))
		return api.GetMonitorLivestatus500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	items := make([]api.LiveStatusItem, len(results))
	for i, r := range results {
		var status string = string(r.Status)
		items[i] = api.LiveStatusItem{
			ServerId:   &r.ServerID,
			ServerName: &r.ServerName,
			Ipv4:       &r.IPv4,
			Status:     &status,
		}
	}

	return api.GetMonitorLivestatus200JSONResponse{
		Items:        &items,
		Total:        &total,
		TotalOnline:  &totalOnline,
		TotalOffline: &totalOffline,
		TotalUnknown: &totalUnknown,
	}, nil
}
