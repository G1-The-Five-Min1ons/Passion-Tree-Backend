package service

import (
	"context"
	"log/slog"

	"passiontree/internal/setting/model"
	"passiontree/internal/setting/repository"
)

type SettingService interface {
	GetSettings(ctx context.Context, userID string) ([]model.Setting, error)
	GetSetting(ctx context.Context, userID, key string) (*model.Setting, error)
	CreateSetting(ctx context.Context, userID string, req *model.SettingRequest) (*model.Setting, error)
	UpdateSetting(ctx context.Context, userID string, req *model.SettingRequest) error
	UpdateMultipleSettings(ctx context.Context, userID string, requests []model.SettingRequest) error
	DeleteSetting(ctx context.Context, userID, key string) error
}

type Service interface {
	SettingService
}

type serviceImpl struct {
	repo   repository.Repository
	logger *slog.Logger
}

func NewService(repo repository.Repository, logger *slog.Logger) Service {
	return &serviceImpl{
		repo:   repo,
		logger: logger,
	}
}
