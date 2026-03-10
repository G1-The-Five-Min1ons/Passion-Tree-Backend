package migration

import (
	"context"
	"fmt"
)

// RunSocialAuthMigration runs the social auth migration for users table.
func (m *Migrator) RunSocialAuthMigration(ctx context.Context) error {
	migrationName := "social_auth_migration"

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

	m.logger.Info("running social auth migration for users table")

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "Check and add auth_provider column",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('users') AND name = 'auth_provider')
				BEGIN
					ALTER TABLE users ADD auth_provider VARCHAR(20) DEFAULT 'local' NOT NULL
				END
			`,
		},
		{
			name: "Check and add provider_user_id column",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('users') AND name = 'provider_user_id')
				BEGIN
					ALTER TABLE users ADD provider_user_id VARCHAR(255) NULL
				END
			`,
		},
		{
			name: "Create provider lookup index",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_users_provider_lookup' AND object_id = OBJECT_ID('users'))
				BEGIN
					CREATE INDEX idx_users_provider_lookup ON users(auth_provider, provider_user_id)
				END
			`,
		},
		{
			name: "Make password nullable",
			sql: `
				IF EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('users') AND name = 'password' AND is_nullable = 0)
				BEGIN
					ALTER TABLE users ALTER COLUMN password VARCHAR(255) NULL
				END
			`,
		},
		{
			name: "Update existing users to local provider",
			sql: `
				UPDATE users SET auth_provider = 'local' WHERE auth_provider IS NULL OR auth_provider = ''
			`,
		},
		{
			name: "Create unique provider index",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_users_provider_unique' AND object_id = OBJECT_ID('users'))
				BEGIN
					CREATE UNIQUE INDEX idx_users_provider_unique ON users(auth_provider, provider_user_id) WHERE provider_user_id IS NOT NULL
				END
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

	m.logger.Info("social auth migration completed successfully", "steps", len(statements))
	return nil
}
