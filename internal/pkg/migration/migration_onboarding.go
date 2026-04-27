package migration

import (
	"context"
	"fmt"
)

// RunOnboardingMigration creates the onboarding_answers table.
func (m *Migrator) RunOnboardingMigration(ctx context.Context) error {
	migrationName := "onboarding_answers_table"

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

	m.logger.Info("running onboarding migration")

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "Create onboarding_answers table",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 'onboarding_answers')
				BEGIN
					CREATE TABLE onboarding_answers (
						id               UNIQUEIDENTIFIER NOT NULL DEFAULT NEWID() PRIMARY KEY,
						user_id          UNIQUEIDENTIFIER NOT NULL,
						subjects         NVARCHAR(500)    NOT NULL,
						knowledge_level  NVARCHAR(100)    NOT NULL,
						motivation       NVARCHAR(200)    NOT NULL,
						daily_goal       NVARCHAR(100)    NOT NULL,
						learning_styles  NVARCHAR(500)    NOT NULL,
						reflection_habit NVARCHAR(100)    NOT NULL,
						created_at       DATETIME2        NOT NULL DEFAULT GETUTCDATE(),
						updated_at       DATETIME2        NOT NULL DEFAULT GETUTCDATE(),
						CONSTRAINT FK_onboarding_answers_users
							FOREIGN KEY (user_id) REFERENCES dbo.Users(user_id) ON DELETE CASCADE,
						CONSTRAINT UQ_onboarding_answers_user UNIQUE (user_id)
					)
				END
			`,
		},
		{
			name: "Create index on user_id",
			sql: `
				IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'IX_onboarding_answers_user_id' AND object_id = OBJECT_ID('onboarding_answers'))
				BEGIN
					CREATE INDEX IX_onboarding_answers_user_id ON onboarding_answers(user_id)
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

	m.logger.Info("onboarding migration completed successfully")
	return nil
}
