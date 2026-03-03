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
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	repository_test "passiontree/internal/tests/unit/learning-path/repository"
)

func TestCreatePath(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreatePathRequest
		setup         func(*repository_test.Repopository)
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
			setup: func(m *repository_test.Repopository) {
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
			setup:         nil,
			expectedError: "title cannot be empty",
		},
		{
			name: "RepositoryError",
			req: model.CreatePathRequest{
				Title:       "Error Path",
				CoverImgURL: "url",
			},
			setup: func(m *repository_test.Repopository) {
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
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			aiClient := aiclient.NewAIClient("http://mock-ai")

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
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGetPaths(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*repository_test.Repopository)
		expectedCount int
		expectedError string
	}{
		{
			name: "Success",
			setup: func(m *repository_test.Repopository) {
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
			setup: func(m *repository_test.Repopository) {
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
			t.Logf("\033[36mExecuting TestGetPaths case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
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
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGetPathDetails(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return &model.LearningPath{PathID: "p1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:   "NotFound",
			pathID: "p2",
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "learning path with id 'p2' not found",
		},
		{
			name:          "EmptyID",
			pathID:        "",
			setup:         nil,
			expectedError: "path_id is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetPathDetails case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetPathDetails(context.Background(), tt.pathID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestUpdatePath(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		req           model.UpdatePathRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			req:    model.UpdatePathRequest{Title: "New Title"},
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return &model.LearningPath{PathID: "p1"}, nil
				}
				m.UpdateLearningPathFunc = func(ctx context.Context, path_id string, req model.UpdatePathRequest) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyPathID",
			pathID:        "",
			req:           model.UpdatePathRequest{Title: "New Title"},
			setup:         nil,
			expectedError: "path_id is required",
		},
		{
			name:   "PathNotFound",
			pathID: "p2",
			req:    model.UpdatePathRequest{Title: "New Title"},
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "cannot update: path_id 'p2' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestUpdatePath case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.UpdatePath(context.Background(), tt.pathID, tt.req)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestDeletePath(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteLearningPathFunc = func(ctx context.Context, path_id string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyPathID",
			pathID:        "",
			setup:         nil,
			expectedError: "path_id is required",
		},
		{
			name:   "NotFound",
			pathID: "p2",
			setup: func(m *repository_test.Repopository) {
				m.DeleteLearningPathFunc = func(ctx context.Context, path_id string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "learning path not found",
		},
		{
			name:   "DependencyConflict",
			pathID: "p3",
			setup: func(m *repository_test.Repopository) {
				m.DeleteLearningPathFunc = func(ctx context.Context, path_id string) error {
					return apperror.NewConflict("foreign key error")
				}
			},
			expectedError: "internal server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestDeletePath case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.DeletePath(context.Background(), tt.pathID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestStartPath(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		userID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			userID: "u1",
			setup: func(m *repository_test.Repopository) {
				m.EnrollLearningPathUserFunc = func(ctx context.Context, pathID, userID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyUserID",
			pathID:        "p1",
			userID:        "",
			setup:         nil,
			expectedError: "user_id is required",
		},
		{
			name:          "EmptyPathID",
			pathID:        "",
			userID:        "u1",
			setup:         nil,
			expectedError: "path_id is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestStartPath case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.StartPath(context.Background(), tt.pathID, tt.userID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGetEnrollmentStatus(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		userID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			userID: "u1",
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathEnrollmentStatusFunc = func(ctx context.Context, pathID, userID string) (*model.PathEnroll, error) {
					return &model.PathEnroll{EnrollID: "enroll-1", Enrollment_status: "in_progress"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:   "NotFound",
			pathID: "p2",
			userID: "u2",
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathEnrollmentStatusFunc = func(ctx context.Context, pathID, userID string) (*model.PathEnroll, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "enrollment not found for user",
		},
		{
			name:          "EmptyParams",
			pathID:        "",
			userID:        "",
			setup:         nil,
			expectedError: "user_id is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetEnrollmentStatus case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetEnrollmentStatus(context.Background(), tt.pathID, tt.userID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGetPathProgress(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		userID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			userID: "u1",
			setup: func(m *repository_test.Repopository) {
				m.GetUserPathProgressFunc = func(ctx context.Context, pathID, userID string) (*model.PathProgressResponse, error) {
					return &model.PathProgressResponse{PathID: pathID, UserID: userID, TotalNodes: 10, CompletedNodes: 5, Progress: 50.0}, nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyParams",
			pathID:        "",
			userID:        "",
			setup:         nil,
			expectedError: "path_id and user_id are required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetPathProgress case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetPathProgress(context.Background(), tt.pathID, tt.userID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGeneratePathWithAI(t *testing.T) {
	t.Run("EmptyTopic", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(nil, nil, logger)
		_, err := svc.GeneratePathWithAI(context.Background(), "")
		if err == nil || !strings.Contains(err.Error(), "topic is required") {
			t.Errorf("Expected 'topic is required' error, got %v", err)
		}
	})

	t.Run("AIFailure", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		aiClient := aiclient.NewAIClient("http://mock-ai") // This will fail network req
		svc := service.NewService(nil, aiClient, logger)
		_, err := svc.GeneratePathWithAI(context.Background(), "Go Lang")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected generation failed error, got %v", err)
		}
	})
}

func TestUpdatePathCoverImage(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		url           string
		expectedError string
	}{
		{
			name:          "EmptyPath",
			pathID:        "",
			url:           "http",
			expectedError: "path_id is required",
		},
		{
			name:          "EmptyURL",
			pathID:        "p1",
			url:           "",
			expectedError: "cover_image_url is required",
		},
		{
			name:          "InvalidURLSource",
			pathID:        "p1",
			url:           "https://example.com/other-folder/img.png",
			expectedError: "Invalid image URL source",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestUpdatePathCoverImage case: %s\033[0m", tt.name)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(nil, nil, logger)

			err := svc.UpdatePathCoverImage(context.Background(), tt.pathID, tt.url)
			if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
			}
		})
	}
}
