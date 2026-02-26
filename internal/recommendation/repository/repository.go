package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/connection"
	"passiontree/internal/recommendation/model"
)

type RepositoryRecommendation interface {
	GetUserReflectionsByTree(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error)
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
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
