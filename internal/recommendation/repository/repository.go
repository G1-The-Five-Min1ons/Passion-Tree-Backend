package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/connection"
	"passiontree/internal/recommendation/model"
)

type RepositoryRecommendation interface {
	GetUserReflectionsByTree(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error)
	GetTopPopularPaths(ctx context.Context) ([]model.RecommendedPath, error)
}

type BatchRepository interface {
	GetBatchInteractions(ctx context.Context) ([]model.UserInteraction, error)
	GetBatchProfiles(ctx context.Context) ([]model.UserProfile, error)
	GetUserInteractions(ctx context.Context, userID string) ([]model.UserInteraction, error)
	GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error)
	SaveBatchRecommendations(ctx context.Context, results []model.BatchRecommendationResult) error
	GetSavedHomeRecommendations(ctx context.Context, userID string) ([]model.RecommendedPath, error)
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Database interface {
	DBTX
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

type Repository interface {
	RepositoryRecommendation
	BatchRepository
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
