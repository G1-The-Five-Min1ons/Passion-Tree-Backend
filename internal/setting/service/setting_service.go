package service

import (
	"context"

	"github.com/google/uuid"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/setting/model"
)

// GetSettings retrieves all settings for a user
func (s *serviceImpl) GetSettings(ctx context.Context, userID string) ([]model.Setting, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	settings, err := s.repo.GetSettings(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get settings", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to retrieve settings: %v", err)
	}

	if settings == nil {
		settings = []model.Setting{}
	}

	return settings, nil
}

// GetSetting retrieves a specific setting
func (s *serviceImpl) GetSetting(ctx context.Context, userID, key string) (*model.Setting, error) {
	if userID == "" || key == "" {
		return nil, apperror.NewBadRequest("user_id and key are required")
	}

	setting, err := s.repo.GetSetting(ctx, userID, key)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get setting", "error", err, "user_id", userID, "key", key)
		return nil, apperror.NewNotFound("setting with key '%s' not found", key)
	}

	return setting, nil
}

// CreateSetting creates a new setting
func (s *serviceImpl) CreateSetting(ctx context.Context, userID string, req *model.SettingRequest) (*model.Setting, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	if req.Key == "" || req.Value == "" {
		return nil, apperror.NewBadRequest("key and value are required")
	}

	// Check if setting already exists
	existing, _ := s.repo.GetSetting(ctx, userID, req.Key)
	if existing != nil {
		return nil, apperror.NewConflict("setting with key '%s' already exists", req.Key)
	}

	setting := &model.Setting{
		SettingID: uuid.New().String(),
		UserID:    userID,
		Key:       req.Key,
		Value:     req.Value,
	}

	err := s.repo.CreateSetting(ctx, setting)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create setting", "error", err, "user_id", userID, "key", req.Key)
		return nil, apperror.NewInternal("failed to create setting: %v", err)
	}

	return setting, nil
}

// UpdateSetting updates an existing setting
func (s *serviceImpl) UpdateSetting(ctx context.Context, userID string, req *model.SettingRequest) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	if req.Key == "" || req.Value == "" {
		return apperror.NewBadRequest("key and value are required")
	}

	err := s.repo.UpdateSetting(ctx, userID, req.Key, req.Value)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update setting", "error", err, "user_id", userID, "key", req.Key)
		return apperror.NewInternal("failed to update setting: %v", err)
	}

	s.logger.InfoContext(ctx, "setting updated successfully", "user_id", userID, "key", req.Key)
	return nil
}

// UpdateMultipleSettings updates multiple settings in a single atomic transaction
func (s *serviceImpl) UpdateMultipleSettings(ctx context.Context, userID string, requests []model.SettingRequest) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	if len(requests) == 0 {
		return apperror.NewBadRequest("at least one setting is required")
	}

	// Validate all requests before proceeding
	keys := make([]string, len(requests))
	values := make([]string, len(requests))
	for i, req := range requests {
		if req.Key == "" || req.Value == "" {
			return apperror.NewBadRequest("setting %d: key and value are required", i+1)
		}
		keys[i] = req.Key
		values[i] = req.Value
	}

	// Call repository method which handles transaction internally
	err := s.repo.UpdateMultipleSettings(ctx, userID, keys, values)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update multiple settings", "error", err, "user_id", userID, "count", len(requests))
		return apperror.NewInternal("failed to update settings: %v", err)
	}

	s.logger.InfoContext(ctx, "batch settings updated successfully", "user_id", userID, "count", len(requests))
	return nil
}

// DeleteSetting deletes a setting
func (s *serviceImpl) DeleteSetting(ctx context.Context, userID, key string) error {
	if userID == "" || key == "" {
		return apperror.NewBadRequest("user_id and key are required")
	}

	err := s.repo.DeleteSetting(ctx, userID, key)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to delete setting", "error", err, "user_id", userID, "key", key)
		return apperror.NewInternal("failed to delete setting: %v", err)
	}

	s.logger.InfoContext(ctx, "setting deleted successfully", "user_id", userID, "key", key)
	return nil
}
