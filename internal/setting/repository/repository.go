package repository

import (
	"context"
	"database/sql"

	"passiontree/internal/connection"
	"passiontree/internal/setting/model"
)

type Repository interface {
	GetSettings(ctx context.Context, userID string) ([]model.Setting, error)
	GetSetting(ctx context.Context, userID, key string) (*model.Setting, error)
	CreateSetting(ctx context.Context, setting *model.Setting) error
	UpdateSetting(ctx context.Context, userID, key, value string) error
	DeleteSetting(ctx context.Context, userID, key string) error
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

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
