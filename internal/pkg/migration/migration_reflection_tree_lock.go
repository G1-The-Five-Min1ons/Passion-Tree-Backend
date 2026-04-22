package migration

import (
	"context"
	"fmt"
)

// RunReflectionTreeLockMigration adds a flag to freeze tree reflection status.
func (m *Migrator) RunReflectionTreeLockMigration(ctx context.Context) error {
	migrationName := "reflection_tree_lock_migration"

	if err := m.EnsureMigrationsTable(ctx); err != nil {
		return err
	}

	ran, err := m.CheckMigrationRan(ctx, migrationName)
	if err != nil {
		return err
	}
	if ran {
		m.logger.Info("migration already applied, skipping", "migration", migrationName)
		return nil
	}

	m.logger.Info("running reflection tree lock migration")

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "Add is_reflection_closed column",
			sql: `
				IF COL_LENGTH('tree', 'is_reflection_closed') IS NULL
				BEGIN
					ALTER TABLE tree ADD is_reflection_closed BIT NOT NULL CONSTRAINT DF_tree_is_reflection_closed DEFAULT 0
				END
			`,
		},
		{
			name: "Backfill existing tree rows",
			sql: `
				UPDATE tree
				SET is_reflection_closed = 0
				WHERE is_reflection_closed IS NULL
			`,
		},
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, stmt := range statements {
		m.logger.Info("executing migration step", "step", stmt.name)
		_, err := tx.ExecContext(ctx, stmt.sql)
		if err != nil {
			m.logger.Error("migration step failed", "step", stmt.name, "error", err)
			return fmt.Errorf("step '%s' failed: %w", stmt.name, err)
		}
	}

	if err := m.RecordMigration(ctx, tx, migrationName); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.logger.Info("reflection tree lock migration completed successfully", "steps", len(statements))
	return nil
}
