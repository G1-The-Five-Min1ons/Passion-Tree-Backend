package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/database"
	"passiontree/internal/reflection/model"
)

type RepositoryReflection interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error)
	GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error)
	GetAllReflections(ctx context.Context) ([]model.Reflection, error)
	UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflection(ctx context.Context, reflectID string) error
	
	// Album methods
	CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (string, error)
	GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error)
	GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error)
	UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error
	DeleteAlbum(ctx context.Context, albumID string) error
	
	// Tree methods
	CreateTree(ctx context.Context, req model.CreateTreeRequest) (string, error)
	GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error)
	GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error)
	UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error
	DeleteTree(ctx context.Context, treeID string) error
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(ds database.Database) RepositoryReflection {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
