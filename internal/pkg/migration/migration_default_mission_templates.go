package migration

import (
	"context"
)

// RunDefaultMissionTemplatesMigration seeds default mission templates.
func (m *Migrator) RunDefaultMissionTemplatesMigration(ctx context.Context) error {
	migrationName := "default_mission_templates"

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

	m.logger.Info("running migration: seed default mission templates")

	query := `
		IF OBJECT_ID('mission', 'U') IS NOT NULL
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM mission WHERE condition_type = 'COMMON_ROUTINE' AND is_active = 1)
			BEGIN
				INSERT INTO mission (
					mission_id, title, detail, reward_xp, reward_heart,
					condition_type, target_value, is_active, create_at, update_at
				)
				VALUES (
					NEWID(), 'Daily Learning Routine', 'Complete at least 1 learning action this week',
					80, 0, 'COMMON_ROUTINE', 1, 1, GETDATE(), GETDATE()
				)
			END

			IF NOT EXISTS (SELECT 1 FROM mission WHERE condition_type = 'COMPLETE_NODE' AND is_active = 1)
			BEGIN
				INSERT INTO mission (
					mission_id, title, detail, reward_xp, reward_heart,
					condition_type, target_value, is_active, create_at, update_at
				)
				VALUES (
					NEWID(), 'Complete Learning Nodes', 'Complete 2 learning nodes',
					120, 1, 'COMPLETE_NODE', 2, 1, GETDATE(), GETDATE()
				)
			END

			IF NOT EXISTS (SELECT 1 FROM mission WHERE condition_type = 'ENROLL_PATH' AND is_active = 1)
			BEGIN
				INSERT INTO mission (
					mission_id, title, detail, reward_xp, reward_heart,
					condition_type, target_value, is_active, create_at, update_at
				)
				VALUES (
					NEWID(), 'Enroll New Path', 'Enroll in 1 new learning path',
					100, 1, 'ENROLL_PATH', 1, 1, GETDATE(), GETDATE()
				)
			END

			IF NOT EXISTS (SELECT 1 FROM mission WHERE condition_type = 'WRITE_REFLECT' AND is_active = 1)
			BEGIN
				INSERT INTO mission (
					mission_id, title, detail, reward_xp, reward_heart,
					condition_type, target_value, is_active, create_at, update_at
				)
				VALUES (
					NEWID(), 'Write Reflection', 'Write 1 reflection this week',
					110, 1, 'WRITE_REFLECT', 1, 1, GETDATE(), GETDATE()
				)
			END
		END

		DECLARE @common_mission_id UNIQUEIDENTIFIER;
		SELECT TOP 1 @common_mission_id = mission_id
		FROM mission
		WHERE condition_type = 'COMMON_ROUTINE' AND is_active = 1

		IF @common_mission_id IS NOT NULL AND OBJECT_ID('user_mission', 'U') IS NOT NULL
		BEGIN
			INSERT INTO user_mission (
				user_mission_id, user_id, mission_id, current_value, status, expire_at
			)
			SELECT
				NEWID(),
				u.user_id,
				@common_mission_id,
				0,
				'active',
				DATEADD(day, 7, CAST(GETDATE() AS DATE))
			FROM users u
			WHERE NOT EXISTS (
				SELECT 1
				FROM user_mission um
				WHERE um.user_id = u.user_id
				  AND um.mission_id = @common_mission_id
			)
		END
	`

	if err := m.RunStatements(ctx, query); err != nil {
		return err
	}

	if err := m.RecordMigration(ctx, nil, migrationName); err != nil {
		return err
	}

	m.logger.Info("default mission templates migration completed successfully")
	return nil
}
