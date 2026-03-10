package repository

import (
	"context"

	"passiontree/internal/setting/model"
)

// GetSettings retrieves all settings for a user
func (r *repositoryImpl) GetSettings(ctx context.Context, userID string) ([]model.Setting, error) {
	query := "SELECT id, user_id, [key], [value], created_at, updated_at FROM settings WHERE user_id = @p1 ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []model.Setting
	for rows.Next() {
		var s model.Setting
		if err := rows.Scan(&s.SettingID, &s.UserID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}

	return settings, rows.Err()
}

// GetSetting retrieves a specific setting
func (r *repositoryImpl) GetSetting(ctx context.Context, userID, key string) (*model.Setting, error) {
	query := "SELECT id, user_id, [key], [value], created_at, updated_at FROM settings WHERE user_id = @p1 AND [key] = @p2"

	var s model.Setting
	err := r.db.QueryRowContext(ctx, query, userID, key).Scan(&s.SettingID, &s.UserID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// CreateSetting creates a new setting
func (r *repositoryImpl) CreateSetting(ctx context.Context, setting *model.Setting) error {
	query := "INSERT INTO settings (id, user_id, [key], [value], created_at, updated_at) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)"

	_, err := r.db.ExecContext(ctx, query, setting.SettingID, setting.UserID, setting.Key, setting.Value, setting.CreatedAt, setting.UpdatedAt)
	return err
}

// UpdateSetting updates an existing setting
func (r *repositoryImpl) UpdateSetting(ctx context.Context, userID, key, value string) error {
	query := `
		IF EXISTS (SELECT 1 FROM settings WHERE user_id = @p2 AND [key] = @p3)
		BEGIN
			UPDATE settings
			SET [value] = @p1, updated_at = GETDATE()
			WHERE user_id = @p2 AND [key] = @p3
		END
		ELSE
		BEGIN
			INSERT INTO settings (id, user_id, [key], [value], created_at, updated_at)
			VALUES (NEWID(), @p2, @p3, @p1, GETDATE(), GETDATE())
		END
	`

	_, err := r.db.ExecContext(ctx, query, value, userID, key)
	return err
}

// DeleteSetting deletes a setting
func (r *repositoryImpl) DeleteSetting(ctx context.Context, userID, key string) error {
	query := "DELETE FROM settings WHERE user_id = @p1 AND [key] = @p2"

	_, err := r.db.ExecContext(ctx, query, userID, key)
	return err
}
