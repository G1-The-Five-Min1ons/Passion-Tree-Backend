package repository

import (
	"context"
	"database/sql"

	"passiontree/internal/connection"
	"passiontree/internal/dashboards/model"
)

type RepositoryDashboard interface {
	GetUserInfo(ctx context.Context, userID string) (*model.UserInfo, error)
	GetWeeklyMissions(ctx context.Context, userID string) ([]model.MissionItem, error)
	GetCurrentPaths(ctx context.Context, userID string) ([]model.CurrentPathItem, error)
	GetUserActivity(ctx context.Context, userID string) ([]model.ActivityItem, error)
	GetActivityHeatmap(ctx context.Context, userID string) ([]model.ActivityHeatmap, error)
	GetTreeCounterStats(ctx context.Context, userID string) (*model.TreeCounterStats, error)
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
	RepositoryDashboard
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
