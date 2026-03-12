package migration

import (
	"context"
	"database/sql"
	"log/slog"
)

// Migrator handles database migrations
// and shared migration execution dependencies.
type Migrator struct {
	db     *sql.DB
	logger *slog.Logger
}

type migrationStep struct {
	name string
	run  func(context.Context) error
}

const migrationTableName = "app_schema_migrations"

// NewMigrator creates a new migration runner.
func NewMigrator(db *sql.DB, logger *slog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}
