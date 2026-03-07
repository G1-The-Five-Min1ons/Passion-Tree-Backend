package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Migrator handles database migrations
type Migrator struct {
	db     *sql.DB
	logger *slog.Logger
}

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

// RunSocialAuthMigration runs the social auth migration for users table
func (m *Migrator) RunSocialAuthMigration(ctx context.Context) error {
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

	for _, stmt := range statements {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := m.db.ExecContext(ctx, stmt.sql)
		cancel()

		if err != nil {
			m.logger.Error("migration step failed",
				"step", stmt.name,
				"error", err,
			)
			return fmt.Errorf("step '%s' failed: %w", stmt.name, err)
		}
	}

	m.logger.Info("social auth migration completed", "steps", len(statements))
	return nil
}

// RunTeacherVerificationAndSettingsMigration ensures teacher verification tables
// and profile settings columns exist.
func (m *Migrator) RunTeacherVerificationAndSettingsMigration(ctx context.Context) error {
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

	for _, stmt := range statements {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := m.db.ExecContext(ctx, stmt.sql)
		cancel()

		if err != nil {
			m.logger.Error("migration step failed",
				"step", stmt.name,
				"error", err,
			)
			return fmt.Errorf("step '%s' failed: %w", stmt.name, err)
		}
	}

	m.logger.Info("teacher verification migration completed", "steps", len(statements))
	return nil
}

// splitByGO splits SQL script by GO keyword (case-insensitive, must be on its own line)
func splitByGO(sql string) []string {
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
