package repository

import (
	"context"
	"database/sql"

	"passiontree/internal/connection"
	"passiontree/internal/onboarding/model"
)

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
	RepositoryOnboarding
}

type RepositoryOnboarding interface {
	UpsertOnboarding(ctx context.Context, userID string, req model.SaveOnboardingRequest) error
	GetOnboardingByUserID(ctx context.Context, userID string) (*model.OnboardingData, error)
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{db: ds.GetDB()}
}
