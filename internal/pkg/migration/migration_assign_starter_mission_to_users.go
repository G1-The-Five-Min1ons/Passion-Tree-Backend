package migration

import (
	"context"
)

// RunAssignStarterMissionToUsersMigration assigns the starter mission to all users.
func (m *Migrator) RunAssignStarterMissionToUsersMigration(ctx context.Context) error {
	migrationName := "assign_starter_mission_to_users"

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

	m.logger.Info("running migration: assign starter mission to all users")

	query := `
		DECLARE @starter_mission_id UNIQUEIDENTIFIER;

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
		END

		SELECT TOP 1 @starter_mission_id = mission_id
		FROM mission
		WHERE condition_type = 'COMMON_ROUTINE' AND is_active = 1;

		IF @starter_mission_id IS NULL
		BEGIN
			RAISERROR('starter mission template not found', 16, 1);
			RETURN;
		END

		IF OBJECT_ID('user_mission', 'U') IS NOT NULL
		BEGIN
			INSERT INTO user_mission (
				user_mission_id, user_id, mission_id, current_value, status, expire_at
			)
			SELECT
				NEWID(),
				u.user_id,
				@starter_mission_id,
				0,
				'active',
				DATEADD(day, 7, CAST(GETDATE() AS DATE))
			FROM users u
			WHERE NOT EXISTS (
				SELECT 1
				FROM user_mission um
				WHERE um.user_id = u.user_id
				  AND um.mission_id = @starter_mission_id
			)
		END
	`

	if err := m.RunStatements(ctx, query); err != nil {
		return err
	}

	if err := m.RecordMigration(ctx, nil, migrationName); err != nil {
		return err
	}

	m.logger.Info("starter mission assignment migration completed successfully")
	return nil
}
