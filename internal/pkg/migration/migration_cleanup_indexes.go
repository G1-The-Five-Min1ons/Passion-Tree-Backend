package migration

import (
	"context"
)

func (m *Migrator) RunCleanupSettingsIndexesMigration(ctx context.Context) error {
    migrationName := "cleanup_redundant_settings_indexes"

    ran, err := m.CheckMigrationRan(ctx, migrationName)
    if err != nil || ran { return err }

    m.logger.Info("running cleanup: dropping redundant indexes on settings table")

    query := `
        IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_settings_user_id' AND object_id = OBJECT_ID('settings'))
            DROP INDEX idx_settings_user_id ON settings;
        
        IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_settings_user_key' AND object_id = OBJECT_ID('settings'))
            DROP INDEX idx_settings_user_key ON settings;
    `

    if err := m.RunStatements(ctx, query); err != nil {
        return err
    }

    return m.RecordMigration(ctx, nil, migrationName)
}