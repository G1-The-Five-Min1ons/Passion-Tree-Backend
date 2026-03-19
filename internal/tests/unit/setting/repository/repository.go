package repository

import (
	"context"

	"passiontree/internal/setting/model"
)

// MockRepo for Setting
type Repository struct {
	GetSettingsFunc            func(ctx context.Context, userID string) ([]model.Setting, error)
	GetSettingFunc             func(ctx context.Context, userID, key string) (*model.Setting, error)
	CreateSettingFunc          func(ctx context.Context, setting *model.Setting) error
	UpdateSettingFunc          func(ctx context.Context, userID, key, value string) error
	UpdateMultipleSettingsFunc func(ctx context.Context, userID string, keys []string, values []string) error
	DeleteSettingFunc          func(ctx context.Context, userID, key string) error
}

func (m *Repository) GetSettings(ctx context.Context, userID string) ([]model.Setting, error) {
	if m.GetSettingsFunc != nil {
		return m.GetSettingsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *Repository) GetSetting(ctx context.Context, userID, key string) (*model.Setting, error) {
	if m.GetSettingFunc != nil {
		return m.GetSettingFunc(ctx, userID, key)
	}
	return nil, nil
}

func (m *Repository) CreateSetting(ctx context.Context, setting *model.Setting) error {
	if m.CreateSettingFunc != nil {
		return m.CreateSettingFunc(ctx, setting)
	}
	return nil
}

func (m *Repository) UpdateSetting(ctx context.Context, userID, key, value string) error {
	if m.UpdateSettingFunc != nil {
		return m.UpdateSettingFunc(ctx, userID, key, value)
	}
	return nil
}

func (m *Repository) UpdateMultipleSettings(ctx context.Context, userID string, keys []string, values []string) error {
	if m.UpdateMultipleSettingsFunc != nil {
		return m.UpdateMultipleSettingsFunc(ctx, userID, keys, values)
	}
	return nil
}

func (m *Repository) DeleteSetting(ctx context.Context, userID, key string) error {
	if m.DeleteSettingFunc != nil {
		return m.DeleteSettingFunc(ctx, userID, key)
	}
	return nil
}
