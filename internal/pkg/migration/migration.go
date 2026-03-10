package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

// Migrator handles database migrations
type Migrator struct {
	db     *sql.DB
	logger *slog.Logger
}

const migrationTableName = "app_schema_migrations"

// NewMigrator creates a new migration runner
func NewMigrator(db *sql.DB, logger *slog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// RunStatements executes SQL statements one by one
// Splits by GO keyword (MSSQL batch separator) or semicolons
func (m *Migrator) RunStatements(ctx context.Context, sqlScript string) error {
	// Split by GO (MSSQL batch separator)
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

// RunSocialAuthMigration runs the social auth migration for users table
// Uses transaction to ensure atomicity - all changes succeed or all rollback
func (m *Migrator) RunSocialAuthMigration(ctx context.Context) error {
	migrationName := "social_auth_migration"

	// Ensure schema_migrations table exists
	if err := m.EnsureMigrationsTable(ctx); err != nil {
		return err
	}

	// Check if migration already ran
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

	// Begin transaction for atomic execution
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Execute all statements within transaction
	for _, stmt := range statements {
		m.logger.Info("executing migration step", "step", stmt.name)

		_, err := tx.ExecContext(ctx, stmt.sql)
		if err != nil {
			m.logger.Error("migration step failed",
				"step", stmt.name,
				"error", err,
			)
			return fmt.Errorf("step '%s' failed: %w", stmt.name, err)
		}
	}

	// Record successful migration
	if err := m.RecordMigration(ctx, tx, migrationName); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.logger.Info("social auth migration completed successfully", "steps", len(statements))
	return nil
}

// RunTeacherVerificationAndSettingsMigration ensures teacher verification tables
// and profile settings columns exist.
// Uses transaction to ensure atomicity - all changes succeed or all rollback
func (m *Migrator) RunTeacherVerificationAndSettingsMigration(ctx context.Context) error {
	migrationName := "teacher_verification_and_settings_migration"

	// Ensure schema_migrations table exists
	if err := m.EnsureMigrationsTable(ctx); err != nil {
		return err
	}

	// Check if migration already ran
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

	// Begin transaction for atomic execution
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Execute all statements within transaction
	for _, stmt := range statements {
		m.logger.Info("executing migration step", "step", stmt.name)

		_, err := tx.ExecContext(ctx, stmt.sql)
		if err != nil {
			m.logger.Error("migration step failed",
				"step", stmt.name,
				"error", err,
			)
			return fmt.Errorf("step '%s' failed: %w", stmt.name, err)
		}
	}

	// Record successful migration
	if err := m.RecordMigration(ctx, tx, migrationName); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.logger.Info("teacher verification migration completed successfully", "steps", len(statements))
	return nil
}

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

// splitByGO splits SQL script by GO keyword (case-insensitive, must be on its own line)
// Handles both Unix (\n) and Windows (\r\n) line endings
func splitByGO(sql string) []string {
	// Normalize line endings to Unix format for consistent handling
	sql = strings.ReplaceAll(sql, "\r\n", "\n")
	lines := strings.Split(sql, "\n")
	var batches []string
	var currentBatch strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, "GO") {
			if currentBatch.Len() > 0 {
				batches = append(batches, currentBatch.String())
				currentBatch.Reset()
			}
		} else {
			currentBatch.WriteString(line)
			currentBatch.WriteString("\n")
		}
	}

	// Add remaining batch
	if currentBatch.Len() > 0 {
		batches = append(batches, currentBatch.String())
	}

	return batches
}

// isCommentOnly checks if a batch contains only comments
func isCommentOnly(batch string) bool {
	lines := strings.Split(batch, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "--") && !strings.HasPrefix(trimmed, "/*") {
			return false
		}
	}
	return true
}

// truncateSQL truncates SQL for logging preview
func truncateSQL(sql string, maxLen int) string {
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")
	if len(sql) > maxLen {
		return sql[:maxLen] + "..."
	}
	return sql
}
