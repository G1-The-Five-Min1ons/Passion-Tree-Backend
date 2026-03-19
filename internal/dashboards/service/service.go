package service

import (
	"context"
	"log/slog"

	"passiontree/internal/dashboards/model"
	"passiontree/internal/dashboards/repository"
)

type ServiceDashboard interface {
	GetDashboardData(ctx context.Context, userID string) (*model.DashboardResponse, error)
}

type Service interface {
	ServiceDashboard
}

type serviceImpl struct {
	dashboardrepo repository.Repository
	logger        *slog.Logger
}

func NewService(repo repository.Repository, logger *slog.Logger) Service {
	return &serviceImpl{
		dashboardrepo: repo,
		logger:        logger,
	}
}
