package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/setting/model"
	"passiontree/internal/setting/service"
	repository_test "passiontree/internal/tests/unit/setting/repository"
)

func TestGetSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingsFunc: func(ctx context.Context, userID string) ([]model.Setting, error) {
				return []model.Setting{
					{SettingID: "s1", UserID: userID, Key: "theme", Value: "dark"},
					{SettingID: "s2", UserID: userID, Key: "language", Value: "th"},
				}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		settings, err := svc.GetSettings(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(settings) != 2 {
			t.Errorf("Expected 2 settings, got %d", len(settings))
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		_, err := svc.GetSettings(context.Background(), "")
		if err == nil || !strings.Contains(err.Error(), "user_id is required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("NilSettings_ReturnsEmptySlice", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingsFunc: func(ctx context.Context, userID string) ([]model.Setting, error) {
				return nil, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		settings, err := svc.GetSettings(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if settings == nil {
			t.Error("Expected empty slice, got nil")
		}
		if len(settings) != 0 {
			t.Errorf("Expected 0 settings, got %d", len(settings))
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingsFunc: func(ctx context.Context, userID string) ([]model.Setting, error) {
				return nil, errors.New("db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		_, err := svc.GetSettings(context.Background(), "user-1")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error, got %v", err)
		}
	})
}

func TestGetSetting(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingFunc: func(ctx context.Context, userID, key string) (*model.Setting, error) {
				return &model.Setting{SettingID: "s1", UserID: userID, Key: key, Value: "dark"}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		setting, err := svc.GetSetting(context.Background(), "user-1", "theme")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if setting == nil || setting.Value != "dark" {
			t.Errorf("Expected setting with value 'dark', got %v", setting)
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		_, err := svc.GetSetting(context.Background(), "", "theme")
		if err == nil || !strings.Contains(err.Error(), "user_id and key are required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("EmptyKey", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		_, err := svc.GetSetting(context.Background(), "user-1", "")
		if err == nil || !strings.Contains(err.Error(), "user_id and key are required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingFunc: func(ctx context.Context, userID, key string) (*model.Setting, error) {
				return nil, errors.New("not found")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		_, err := svc.GetSetting(context.Background(), "user-1", "unknown")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})
}

func TestCreateSetting(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingFunc: func(ctx context.Context, userID, key string) (*model.Setting, error) {
				return nil, errors.New("not found")
			},
			CreateSettingFunc: func(ctx context.Context, setting *model.Setting) error {
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		req := &model.SettingRequest{Key: "theme", Value: "dark"}
		setting, err := svc.CreateSetting(context.Background(), "user-1", req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if setting == nil || setting.Key != "theme" || setting.Value != "dark" {
			t.Errorf("Expected setting with key 'theme', got %v", setting)
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		_, err := svc.CreateSetting(context.Background(), "", &model.SettingRequest{Key: "k", Value: "v"})
		if err == nil || !strings.Contains(err.Error(), "user_id is required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("EmptyKey", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		_, err := svc.CreateSetting(context.Background(), "user-1", &model.SettingRequest{Key: "", Value: "v"})
		if err == nil || !strings.Contains(err.Error(), "key and value are required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("EmptyValue", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		_, err := svc.CreateSetting(context.Background(), "user-1", &model.SettingRequest{Key: "k", Value: ""})
		if err == nil || !strings.Contains(err.Error(), "key and value are required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingFunc: func(ctx context.Context, userID, key string) (*model.Setting, error) {
				return &model.Setting{SettingID: "s1", Key: key}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		_, err := svc.CreateSetting(context.Background(), "user-1", &model.SettingRequest{Key: "theme", Value: "dark"})
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected conflict error, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetSettingFunc: func(ctx context.Context, userID, key string) (*model.Setting, error) {
				return nil, errors.New("not found")
			},
			CreateSettingFunc: func(ctx context.Context, setting *model.Setting) error {
				return errors.New("db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		_, err := svc.CreateSetting(context.Background(), "user-1", &model.SettingRequest{Key: "theme", Value: "dark"})
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error, got %v", err)
		}
	})
}

func TestUpdateSetting(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateSettingFunc: func(ctx context.Context, userID, key, value string) error {
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		err := svc.UpdateSetting(context.Background(), "user-1", &model.SettingRequest{Key: "theme", Value: "light"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		err := svc.UpdateSetting(context.Background(), "", &model.SettingRequest{Key: "k", Value: "v"})
		if err == nil || !strings.Contains(err.Error(), "user_id is required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("EmptyKeyOrValue", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		err := svc.UpdateSetting(context.Background(), "user-1", &model.SettingRequest{Key: "", Value: "v"})
		if err == nil || !strings.Contains(err.Error(), "key and value are required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateSettingFunc: func(ctx context.Context, userID, key, value string) error {
				return errors.New("db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		err := svc.UpdateSetting(context.Background(), "user-1", &model.SettingRequest{Key: "theme", Value: "light"})
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error, got %v", err)
		}
	})
}

func TestUpdateMultipleSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateMultipleSettingsFunc: func(ctx context.Context, userID string, keys []string, values []string) error {
				if len(keys) != 2 || len(values) != 2 {
					t.Fatalf("Expected 2 keys and values, got %d and %d", len(keys), len(values))
				}
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		requests := []model.SettingRequest{
			{Key: "theme", Value: "dark"},
			{Key: "language", Value: "en"},
		}
		err := svc.UpdateMultipleSettings(context.Background(), "user-1", requests)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		err := svc.UpdateMultipleSettings(context.Background(), "", []model.SettingRequest{{Key: "k", Value: "v"}})
		if err == nil || !strings.Contains(err.Error(), "user_id is required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("EmptyRequests", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		err := svc.UpdateMultipleSettings(context.Background(), "user-1", []model.SettingRequest{})
		if err == nil || !strings.Contains(err.Error(), "at least one setting is required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("InvalidRequestItem", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		requests := []model.SettingRequest{
			{Key: "theme", Value: "dark"},
			{Key: "", Value: "en"},
		}
		err := svc.UpdateMultipleSettings(context.Background(), "user-1", requests)
		if err == nil || !strings.Contains(err.Error(), "key and value are required") {
			t.Errorf("Expected validation error for item, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateMultipleSettingsFunc: func(ctx context.Context, userID string, keys []string, values []string) error {
				return errors.New("db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		requests := []model.SettingRequest{{Key: "theme", Value: "dark"}}
		err := svc.UpdateMultipleSettings(context.Background(), "user-1", requests)
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error, got %v", err)
		}
	})
}

func TestDeleteSetting(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			DeleteSettingFunc: func(ctx context.Context, userID, key string) error {
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		err := svc.DeleteSetting(context.Background(), "user-1", "theme")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		err := svc.DeleteSetting(context.Background(), "", "theme")
		if err == nil || !strings.Contains(err.Error(), "user_id and key are required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("EmptyKey", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		err := svc.DeleteSetting(context.Background(), "user-1", "")
		if err == nil || !strings.Contains(err.Error(), "user_id and key are required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			DeleteSettingFunc: func(ctx context.Context, userID, key string) error {
				return errors.New("db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		err := svc.DeleteSetting(context.Background(), "user-1", "theme")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error, got %v", err)
		}
	})
}
