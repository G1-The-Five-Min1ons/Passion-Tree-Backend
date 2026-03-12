package migration

import (
	"context"
	"fmt"
)

// RunTeacherVerificationAndSettingsMigration ensures teacher verification tables and profile settings columns exist.
func (m *Migrator) RunTeacherVerificationAndSettingsMigration(ctx context.Context) error {
	migrationName := "teacher_verification_and_settings_migration"

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

	m.logger.Info("running teacher verification and profile settings migration")

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "Add profile phone number column",
			sql: `
				IF COL_LENGTH('profile', 'Phone_Number') IS NULL
				BEGIN
					ALTER TABLE profile ADD Phone_Number NVARCHAR(20) NULL
				END
			`,
		},
		{
			name: "Add profile time zone column",
			sql: `
				IF COL_LENGTH('profile', 'Time_Zone') IS NULL
				BEGIN
					ALTER TABLE profile ADD Time_Zone NVARCHAR(64) NOT NULL CONSTRAINT DF_profile_time_zone DEFAULT 'Asia/Bangkok'
				END
			`,
		},
		{
			name: "Add profile date format column",
			sql: `
				IF COL_LENGTH('profile', 'Date_Format') IS NULL
				BEGIN
					ALTER TABLE profile ADD Date_Format NVARCHAR(32) NOT NULL CONSTRAINT DF_profile_date_format DEFAULT 'DD/MM/YYYY'
				END
			`,
		},
		{
			name: "Create teacher verification table",
			sql: `
				IF OBJECT_ID('teacher_verification_requests', 'U') IS NULL
				BEGIN
					CREATE TABLE teacher_verification_requests (
						request_id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT NEWID(),
						user_id UNIQUEIDENTIFIER NOT NULL,
						phone_number NVARCHAR(20) NOT NULL,
						reason NVARCHAR(500) NOT NULL,
						teaching_history NVARCHAR(MAX) NOT NULL,
						status NVARCHAR(20) NOT NULL DEFAULT 'pending',
						reviewed_by UNIQUEIDENTIFIER NULL,
						reviewed_at DATETIME NULL,
						created_at DATETIME NOT NULL DEFAULT GETDATE(),
						updated_at DATETIME NOT NULL DEFAULT GETDATE(),
						CONSTRAINT FK_teacher_verification_user FOREIGN KEY (user_id) REFERENCES users(user_id),
						CONSTRAINT FK_teacher_verification_reviewer FOREIGN KEY (reviewed_by) REFERENCES users(user_id),
						CONSTRAINT UQ_teacher_verification_user UNIQUE (user_id)
					)
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

	m.logger.Info("teacher verification migration completed successfully", "steps", len(statements))
	return nil
}
