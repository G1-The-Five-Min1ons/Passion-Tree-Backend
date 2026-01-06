package service

import (
	"context"
	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
)

type ReflectionService interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error)
	GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error)
	GetAllReflections(ctx context.Context) ([]model.Reflection, error)
	UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflection(ctx context.Context, reflectID string) error
}

type serviceImpl struct {
	refRepo repository.RepositoryReflection
}

func NewService(repo repository.RepositoryReflection) ReflectionService {
	return &serviceImpl{
		refRepo: repo,
	}
}
