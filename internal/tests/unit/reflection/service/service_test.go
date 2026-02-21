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
		mockSetup     func(*repository_test.MockRefRepo)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return &model.Reflection{ReflectID: "r1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "reflection with id 'r2' not found",
		},
		{
			name:      "InternalError",
			reflectID: "r3",
			mockSetup: func(m *repository_test.MockRefRepo) {
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
			mock := &repository_test.MockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger) // aiClient nil

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

func TestDeleteReflection(t *testing.T) {
	tests := []struct {
		name          string
		reflectID     string
		mockSetup     func(*repository_test.MockRefRepo)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *repository_test.MockRefRepo) {
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
			mock := &repository_test.MockRefRepo{}
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
	// Test that it fails when AI service is not configured (nil aiClient)
	mock := &repository_test.MockRefRepo{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewService(mock, nil, logger)

	req := model.CreateReflectionRequest{
		LearningReflect: "I learned a lot",
		MoodReflect:     "Happy",
		TreeNodeID:      "node-1",
	}

	_, err := svc.CreateReflection(context.Background(), req)
	expectedError := "internal server error"
	if err == nil {
		t.Errorf("Expected error '%s', got nil", expectedError)
	} else if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCreateReflection_MissingFields(t *testing.T) {
	t.Log("\033[36mExecuting TestCreateReflection_MissingFields\033[0m")
	mock := &repository_test.MockRefRepo{}
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
