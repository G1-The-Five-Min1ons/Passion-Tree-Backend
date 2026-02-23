package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/service"
	repository_test "passiontree/internal/tests/unit/reflection/repository"
)

func TestGetReflectionByID(t *testing.T) {
	tests := []struct {
		name          string
		reflectID     string
		mockSetup     func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *repository_test.Repository) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return &model.Reflection{ReflectID: "r1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *repository_test.Repository) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "reflection with id 'r2' not found",
		},
		{
			name:      "InternalError",
			reflectID: "r3",
			mockSetup: func(m *repository_test.Repository) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return nil, errors.New("db fail")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetReflectionByID case: %s\033[0m", tt.name)
			mock := &repository_test.Repository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetReflectionByID(context.Background(), tt.reflectID)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

func TestGetAllReflections(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetAllReflectionsFunc: func(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
				return []model.Reflection{{ReflectID: "r1"}}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetAllReflections(context.Background(), model.GetReflectionsFilter{})
		if err != nil || len(res) == 0 {
			t.Errorf("Expected valid reflections list, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetAllReflectionsFunc: func(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
				return nil, errors.New("db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetAllReflections(context.Background(), model.GetReflectionsFilter{})
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected an internal server error, got %v", err)
		}
	})
}

func TestUpdateReflection(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateReflectionFunc: func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateReflectionRequest{LearningReflect: "Test", MoodReflect: "Happy"}
		err := svc.UpdateReflection(context.Background(), "r1", req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(nil, nil, logger)

		err := svc.UpdateReflection(context.Background(), "ref-1", model.UpdateReflectionRequest{})
		if err == nil || !strings.Contains(err.Error(), "learning_reflect is required") {
			t.Errorf("Expected empty body error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateReflectionFunc: func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
				return sql.ErrNoRows
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateReflectionRequest{LearningReflect: "Test", MoodReflect: "Happy"}
		err := svc.UpdateReflection(context.Background(), "r2", req)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})
}

func TestDeleteReflection(t *testing.T) {
	tests := []struct {
		name          string
		reflectID     string
		mockSetup     func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *repository_test.Repository) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *repository_test.Repository) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "reflection with id 'r2' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestDeleteReflection case: %s\033[0m", tt.name)
			mock := &repository_test.Repository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.DeleteReflection(context.Background(), tt.reflectID)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

func TestCreateReflection_NoAI(t *testing.T) {
	t.Log("\033[36mExecuting TestCreateReflection_NoAI\033[0m")
	mock := &repository_test.Repository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewService(mock, nil, logger)

	req := model.CreateReflectionRequest{
		LearningReflect: "I learned a lot",
		MoodReflect:     "Happy",
		TreeNodeID:      "node-1",
	}

	_, err := svc.CreateReflection(context.Background(), req)
	expectedError := "internal server error" // fails when AI is missing
	if err == nil {
		t.Errorf("Expected error '%s', got nil", expectedError)
	} else if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCreateReflection_MissingFields(t *testing.T) {
	t.Log("\033[36mExecuting TestCreateReflection_MissingFields\033[0m")
	mock := &repository_test.Repository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewService(mock, nil, logger)

	req := model.CreateReflectionRequest{
		// Missing fields
	}

	_, err := svc.CreateReflection(context.Background(), req)
	expectedError := "learning_reflect is required"
	if err == nil {
		t.Errorf("Expected error '%s', got nil", expectedError)
	} else if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}
