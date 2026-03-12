package migration

import (
	"context"
	"fmt"
)

// RunPlatformAnnouncementsMigration creates the table for platform-wide announcements used in email notifications.
func (m *Migrator) RunPlatformAnnouncementsMigration(ctx context.Context) error {
	migrationName := "platform_announcements_migration"

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

	m.logger.Info("running platform announcements migration")

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "Create platform_announcements table",
			sql: `
				IF OBJECT_ID('platform_announcements', 'U') IS NULL
				BEGIN
					CREATE TABLE platform_announcements (
						id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT NEWID(),
						title NVARCHAR(200) NOT NULL,
						content NVARCHAR(MAX) NOT NULL,
						is_active BIT NOT NULL DEFAULT 1,
						publish_at DATETIME NOT NULL DEFAULT GETDATE(),
						created_at DATETIME NOT NULL DEFAULT GETDATE(),
						updated_at DATETIME NOT NULL DEFAULT GETDATE()
					)
				END
			`,
		},
		{
			name: "Create publish and active index",
			sql: `
				IF NOT EXISTS (
					SELECT 1
					FROM sys.indexes
					WHERE name = 'idx_platform_announcements_active_publish'
					  AND object_id = OBJECT_ID('platform_announcements')
				)
				BEGIN
					CREATE INDEX idx_platform_announcements_active_publish
					ON platform_announcements(is_active, publish_at DESC)
				END
			`,
		},
		{
			name: "Create updated_at trigger",
			sql: `
				IF OBJECT_ID('trg_platform_announcements_updated_at', 'TR') IS NULL
				BEGIN
					EXEC('
						CREATE TRIGGER trg_platform_announcements_updated_at
						ON platform_announcements
						AFTER UPDATE
						AS
						BEGIN
							SET NOCOUNT ON;
							UPDATE pa
							SET updated_at = GETDATE()
							FROM platform_announcements pa
							INNER JOIN inserted i ON pa.id = i.id;
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
		if _, err := tx.ExecContext(ctx, stmt.sql); err != nil {
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

	m.logger.Info("platform announcements migration completed successfully", "steps", len(statements))
	return nil
}
