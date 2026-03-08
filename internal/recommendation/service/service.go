package service

import (
	"context"
	"log/slog"

	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"

	pathrepo "passiontree/internal/learning-path/repository"
	recrepo "passiontree/internal/recommendation/repository"
)

type ServiceRecommendation interface {
	RecommendPathsForUser(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error)
	RecommendHomePathsForUser(ctx context.Context, userID string) (*model.RecommendPathResponse, error)
	getFallbackPopularPaths(ctx context.Context, message string) (*model.RecommendPathResponse, error)
	extractPathID(id any) (string, bool)
}

type Service interface {
	ServiceRecommendation
}

type serviceImpl struct {
	recRepo  recrepo.Repository
	pathRepo pathrepo.RepositoryLearningPath
	aiClient *aiclient.AIClient
	logger   *slog.Logger
}

func NewService(recRepo recrepo.Repository, pathRepo pathrepo.RepositoryLearningPath, aiClient *aiclient.AIClient, logger *slog.Logger) Service {
	return &serviceImpl{
		recRepo:  recRepo,
		pathRepo: pathRepo,
		aiClient: aiClient,
		logger:   logger,
	}
}
