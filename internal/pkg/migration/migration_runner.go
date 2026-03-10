package migration

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// RunStatements executes SQL statements one by one
// Splits by GO keyword (MSSQL batch separator) or semicolons
func (m *Migrator) RunStatements(ctx context.Context, sqlScript string) error {
	batches := splitByGO(sqlScript)

	m.logger.Info("starting migration", "total_batches", len(batches))

	for i, batch := range batches {
		batch = strings.TrimSpace(batch)
		if batch == "" || isCommentOnly(batch) {
			continue
		}

		m.logger.Info("executing batch", "batch_number", i+1)

		_, err := m.db.ExecContext(ctx, batch)
		if err != nil {
			m.logger.Error("migration batch failed",
				"batch_number", i+1,
				"error", err,
				"sql_preview", truncateSQL(batch, 100),
			)
			return fmt.Errorf("batch %d failed: %w", i+1, err)
		}

		m.logger.Info("batch executed successfully", "batch_number", i+1)
	}

	m.logger.Info("migration completed successfully")
	return nil
}

// EnsureMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) EnsureMigrationsTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		IF OBJECT_ID('%s', 'U') IS NULL
		BEGIN
			CREATE TABLE %s (
				id INT IDENTITY(1,1) PRIMARY KEY,
				migration_name NVARCHAR(255) NOT NULL UNIQUE,
				applied_at DATETIME NOT NULL DEFAULT GETDATE()
			)
		END
	`, migrationTableName, migrationTableName)

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// CheckMigrationRan checks if a specific migration has already been applied
func (m *Migrator) CheckMigrationRan(ctx context.Context, migrationName string) (bool, error) {
	var count int
	err := m.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE migration_name = @p1", migrationTableName),
		migrationName,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	return count > 0, nil
}

// RecordMigration records a successful migration in the schema_migrations table
func (m *Migrator) RecordMigration(ctx context.Context, tx *sql.Tx, migrationName string) error {
	var executor interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}

	if tx != nil {
		executor = tx
	} else {
		executor = m.db
	}

	_, err := executor.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (migration_name, applied_at) VALUES (@p1, GETDATE())", migrationTableName),
		migrationName,
	)

	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}
