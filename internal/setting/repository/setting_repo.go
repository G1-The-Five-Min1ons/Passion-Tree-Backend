package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"passiontree/internal/setting/model"

	"github.com/google/uuid"
)

// GetSettings retrieves all settings for a user
func (r *repositoryImpl) GetSettings(ctx context.Context, userID string) ([]model.Setting, error) {
	query := "SELECT CONVERT(VARCHAR(36), id) AS id, CONVERT(VARCHAR(36), user_id) AS user_id, [key], [value], created_at, updated_at FROM settings WHERE user_id = @p1 ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query settings failed for user %s: %w", userID, err)
	}
	defer rows.Close()

	var settings []model.Setting
	for rows.Next() {
		var s model.Setting
		if err := rows.Scan(&s.SettingID, &s.UserID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan setting row failed: %w", err)
		}
		settings = append(settings, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return settings, rows.Err()
}

// GetSetting retrieves a specific setting
func (r *repositoryImpl) GetSetting(ctx context.Context, userID, key string) (*model.Setting, error) {
	query := "SELECT CONVERT(VARCHAR(36), id) AS id, CONVERT(VARCHAR(36), user_id) AS user_id, [key], [value], created_at, updated_at FROM settings WHERE user_id = @p1 AND [key] = @p2"

	var s model.Setting
	err := r.db.QueryRowContext(ctx, query, userID, key).Scan(&s.SettingID, &s.UserID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get setting failed (user: %s, key: %s): %w", userID, key, err)
	}
	return &s, nil
}

// CreateSetting creates a new setting
func (r *repositoryImpl) CreateSetting(ctx context.Context, setting *model.Setting) error {
	query := "INSERT INTO settings (id, user_id, [key], [value], created_at, updated_at) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)"

	_, err := r.db.ExecContext(ctx, query, 
        setting.SettingID, 
        setting.UserID, 
        setting.Key, 
        setting.Value, 
        setting.CreatedAt, 
        setting.UpdatedAt,
    )
	if err != nil {
        return fmt.Errorf("create setting failed (user_id: %s, key: %s): %w", setting.UserID, setting.Key, err)
    }

	return nil
}

// UpdateSetting updates an existing setting
func (r *repositoryImpl) UpdateSetting(ctx context.Context, userID, key, value string) error {
	newID := uuid.New().String()
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
			VALUES (@p4, @p2, @p3, @p1, GETDATE(), GETDATE())
		END
	`

	_, err := r.db.ExecContext(ctx, query, value, userID, key, newID)

	if err != nil {
		return fmt.Errorf("update setting failed (user_id: %s, key: %s): %w", userID, key, err)
	}

	return nil
}

// DeleteSetting deletes a setting
func (r *repositoryImpl) DeleteSetting(ctx context.Context, userID, key string) error {
	query := "DELETE FROM settings WHERE user_id = @p1 AND [key] = @p2"

	result, err := r.db.ExecContext(ctx, query, userID, key)
	if err != nil {
        return fmt.Errorf("delete setting failed (user_id: %s, key: %s): %w", userID, key, err)
    }

	rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("could not get rows affected: %w", err)
    }

	if rows == 0 {
		return fmt.Errorf("setting not found (user_id: %s, key: %s)", userID, key)
    }

	return nil
}

// UpdateMultipleSettings updates multiple settings in a single atomic transaction
func (r *repositoryImpl) UpdateMultipleSettings(ctx context.Context, userID string, keys []string, values []string) error {
	if len(keys) == 0 {
		return nil
	}

	items := make([]model.SettingItem, len(keys))
	for i := 0; i < len(keys); i++ {
		items[i] = model.SettingItem{
			ID:    uuid.New().String(),
			Key:   keys[i],
			Value: values[i],
		}
	}

	jsonData, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
        MERGE INTO settings AS target
        USING (
            SELECT id, [key], [value]
            FROM OPENJSON(@p1)
            WITH (id NVARCHAR(50), [key] NVARCHAR(255), [value] NVARCHAR(MAX))
        ) AS source
        ON (target.user_id = @p2 AND target.[key] = source.[key])
        WHEN MATCHED THEN
            UPDATE SET [value] = source.[value], updated_at = GETDATE()
        WHEN NOT MATCHED THEN
            INSERT (id, user_id, [key], [value], created_at, updated_at)
            VALUES (source.id, @p2, source.[key], source.[value], GETDATE(), GETDATE());
    `

	_, err = r.db.ExecContext(ctx, query, string(jsonData), userID)
	if err != nil {
		return fmt.Errorf("batch update settings failed: %w", err)
	}

	return nil
}
