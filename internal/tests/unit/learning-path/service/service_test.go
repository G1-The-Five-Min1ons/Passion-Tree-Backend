package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/platform/aiclient"
	repository_test "passiontree/internal/tests/unit/learning-path/repository"
)

func TestCreatePath(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreatePathRequest
		mockSetup     func(*repository_test.MockRepo)
		expectedError string
		expectedID    string
	}{
		{
			name: "Success",
			req: model.CreatePathRequest{
				Title:       "Go Lang",
				CoverImgURL: "https://example.com/learning-path/go.png",
				CreatorID:   "user-1",
			},
			mockSetup: func(m *repository_test.MockRepo) {
				m.CreateLearningPathFunc = func(ctx context.Context, req model.CreatePathRequest) (string, error) {
					return "path-123", nil
				}
			},
			expectedID:    "path-123",
			expectedError: "",
		},
		{
			name: "MissingTitle",
			req: model.CreatePathRequest{
				Title:       "",
				CoverImgURL: "url",
			},
			mockSetup:     nil,
			expectedError: "title cannot be empty",
		},
		{
			name: "RepositoryError",
			req: model.CreatePathRequest{
				Title:       "Error Path",
				CoverImgURL: "url",
			},
			mockSetup: func(m *repository_test.MockRepo) {
				m.CreateLearningPathFunc = func(ctx context.Context, req model.CreatePathRequest) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestCreatePath case: %s\033[0m", tt.name)
			mock := &repository_test.MockRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			aiClient := aiclient.NewAIClient("http://mock-ai")

			// Use service.NewService
			svc := service.NewService(mock, aiClient, logger)

			id, err := svc.CreatePath(context.Background(), tt.req)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if id != tt.expectedID {
					t.Errorf("Expected ID %s, got %s", tt.expectedID, id)
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

func TestGetPaths(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*repository_test.MockRepo)
		expectedCount int
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func(m *repository_test.MockRepo) {
				m.GetAllLearningPathFunc = func(ctx context.Context) ([]model.LearningPath, error) {
					return []model.LearningPath{
						{PathID: "1", Title: "Path 1"},
						{PathID: "2", Title: "Path 2"},
					}, nil
				}
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name: "Error",
			mockSetup: func(m *repository_test.MockRepo) {
				m.GetAllLearningPathFunc = func(ctx context.Context) ([]model.LearningPath, error) {
					return nil, errors.New("db fail")
				}
			},
			expectedCount: 0,
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestCreatePath case: %s\033[0m", tt.name)
			mock := &repository_test.MockRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			// Use service.NewService (nil AI client)
			svc := service.NewService(mock, nil, logger)

			paths, err := svc.GetPaths(context.Background())

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(paths) != tt.expectedCount {
					t.Errorf("Expected %d paths, got %d", tt.expectedCount, len(paths))
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

func TestGetPathDetails(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		mockSetup     func(*repository_test.MockRepo)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			mockSetup: func(m *repository_test.MockRepo) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return &model.LearningPath{PathID: "p1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:   "NotFound",
			pathID: "p2",
			mockSetup: func(m *repository_test.MockRepo) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return nil, sql.ErrNoRows
				}
			},
			// "learning path with id '%s' not found"
			expectedError: "learning path with id 'p2' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetPaths case: %s\033[0m", tt.name)
			mock := &repository_test.MockRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			// Use service.NewService (nil AI client)
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetPathDetails(context.Background(), tt.pathID)
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
