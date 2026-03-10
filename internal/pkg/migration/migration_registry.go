package migration

import (
	"context"
	"fmt"
)

// RunAllMigrations executes all application migrations in deterministic order.
func (m *Migrator) RunAllMigrations(ctx context.Context) error {
	steps := []migrationStep{
		{name: "social_auth", run: m.RunSocialAuthMigration},
		{name: "teacher_verification_and_settings", run: m.RunTeacherVerificationAndSettingsMigration},
		{name: "user_settings_table", run: m.RunUserSettingsTableMigration},
		{name: "remove_web_push_setting", run: m.RunRemoveWebPushSettingMigration},
	}

	for _, step := range steps {
		m.logger.Info("running migration step", "step", step.name)
		if err := step.run(ctx); err != nil {
			return fmt.Errorf("migration step '%s' failed: %w", step.name, err)
		}
	}

	return nil
}
