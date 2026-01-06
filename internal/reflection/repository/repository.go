package repository

import (
	"context"
	"database/sql"

	"passiontree/internal/database"
	"passiontree/internal/reflection/model"
)

type RepositoryReflection interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (string, error)
	GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error)
	GetAllReflections(ctx context.Context) ([]model.Reflection, error)
	UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflection(ctx context.Context, reflectID string) error
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(ds database.Database) RepositoryReflection {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
