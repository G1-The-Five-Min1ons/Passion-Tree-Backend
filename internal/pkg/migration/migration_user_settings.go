package migration

import (
	"context"
	"fmt"
)

// RunUserSettingsTableMigration ensures settings table exists for key-value user settings.
func (m *Migrator) RunUserSettingsTableMigration(ctx context.Context) error {
	migrationName := "user_settings_table_migration"

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

	m.logger.Info("running user settings table migration")

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "Create settings table",
			sql: `
				IF OBJECT_ID('settings', 'U') IS NULL
				BEGIN
					CREATE TABLE settings (
						id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT NEWID(),
						user_id UNIQUEIDENTIFIER NOT NULL,
						[key] NVARCHAR(100) NOT NULL,
						[value] NVARCHAR(MAX) NOT NULL,
						created_at DATETIME NOT NULL DEFAULT GETDATE(),
						updated_at DATETIME NOT NULL DEFAULT GETDATE(),
						CONSTRAINT FK_settings_user FOREIGN KEY (user_id) REFERENCES users(user_id),
						CONSTRAINT UQ_settings_user_key UNIQUE (user_id, [key])
					)
				END
			`,
		},
		{
			name: "Create settings user index",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_settings_user_id' AND object_id = OBJECT_ID('settings'))
				BEGIN
					CREATE INDEX idx_settings_user_id ON settings(user_id)
				END
			`,
		},
		{
			name: "Create settings user+key index",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_settings_user_key' AND object_id = OBJECT_ID('settings'))
				BEGIN
					CREATE INDEX idx_settings_user_key ON settings(user_id, [key])
				END
			`,
		},
		{
			name: "Create settings updated_at trigger",
			sql: `
				IF OBJECT_ID('trg_settings_updated_at', 'TR') IS NULL
				BEGIN
					EXEC('
						CREATE TRIGGER trg_settings_updated_at
						ON settings
						AFTER UPDATE
						AS
						BEGIN
							SET NOCOUNT ON;
							UPDATE s
							SET updated_at = GETDATE()
							FROM settings s
							INNER JOIN inserted i ON s.id = i.id;
						END
					')
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

	m.logger.Info("user settings table migration completed successfully", "steps", len(statements))
	return nil
}
