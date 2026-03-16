package migration

import (
	"context"
	"fmt"
)

// RunRemoveWebPushSettingMigration removes deprecated web push setting entries.
// Safe to run once; statement is idempotent because DELETE on missing rows is a no-op.
func (m *Migrator) RunRemoveWebPushSettingMigration(ctx context.Context) error {
	migrationName := "remove_web_push_setting_migration"

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

	m.logger.Info("running remove web push setting migration")

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM settings WHERE [key] = 'web_push_enabled'`); err != nil {
		m.logger.Error("migration step failed", "step", "delete deprecated web push setting", "error", err)
		return fmt.Errorf("step 'delete deprecated web push setting' failed: %w", err)
	}

	if err := m.RecordMigration(ctx, tx, migrationName); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.logger.Info("remove web push setting migration completed successfully")
	return nil
}
