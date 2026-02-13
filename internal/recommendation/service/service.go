package service

import (
	"context"
	"log/slog"
	"passiontree/internal/recommendation/model"
	"passiontree/internal/recommendation/repository"
)

type ServiceRecommendation interface {
	RecommendPathsForUser(ctx context.Context, userID string) (*model.RecommendPathResponse, error)
}

type Service interface {
	ServiceRecommendation
}

type serviceImpl struct {
	recreflectRepo    repository.RepositoryRecommendation
	logger  	*slog.Logger
}

func NewService(repo repository.Repository, logger *slog.Logger) Service {
	return &serviceImpl{
		recreflectRepo:    repo,
		logger:   	 logger,
	}
}
