package service

import (
	"context"
	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
	"passiontree/internal/platform/aiclient"
)

type ReflectionService interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error)
	GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error)
	GetAllReflections(ctx context.Context) ([]model.Reflection, error)
	UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflection(ctx context.Context, reflectID string) error
}

type serviceImpl struct {
	refRepo  repository.RepositoryReflection
	aiClient *aiclient.AIClient
}

func NewService(repo repository.RepositoryReflection, aiClient *aiclient.AIClient) ReflectionService {
	return &serviceImpl{
		refRepo:  repo,
		aiClient: aiClient,
	}
}