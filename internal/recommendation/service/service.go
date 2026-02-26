package service

import (
	"context"
	"log/slog"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"
	"passiontree/internal/recommendation/repository"
)

type ServiceRecommendation interface {
	RecommendPathsForUser(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error)
}

type Service interface {
	ServiceRecommendation
}

type serviceImpl struct {
	recreflectRepo repository.RepositoryRecommendation
	aiClient       *aiclient.AIClient
	logger         *slog.Logger
}

func NewService(repo repository.Repository, aiClient *aiclient.AIClient, logger *slog.Logger) Service {
	return &serviceImpl{
		recreflectRepo: repo,
		aiClient:       aiClient,
		logger:         logger,
	}
}
