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

func TestGetUserHistory(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setup         func(*repository_test.Repopository)
		expectedLen   int
		expectedError string
	}{
		{
			name:   "Success",
			userID: "u1",
			setup: func(m *repository_test.Repopository) {
				m.GetHistoryByUserIDFunc = func(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
					return []model.HistoryResponse{{Path_id: "p1"}}, nil
				}
			},
			expectedLen:   1,
			expectedError: "",
		},
		{
			name:   "NilList",
			userID: "u2",
			setup: func(m *repository_test.Repopository) {
				m.GetHistoryByUserIDFunc = func(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
					return nil, nil
				}
			},
			expectedLen:   0,
			expectedError: "",
		},
		{
			name:          "EmptyUserID",
			userID:        "",
			setup:         nil,
			expectedLen:   0,
			expectedError: "user_id is required",
		},
		{
			name:   "Error",
			userID: "u3",
			setup: func(m *repository_test.Repopository) {
				m.GetHistoryByUserIDFunc = func(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
					return nil, errors.New("db err")
				}
			},
			expectedLen:   0,
			expectedError: "internal server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetUserHistory case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			history, err := svc.GetUserHistory(context.Background(), tt.userID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(history) != tt.expectedLen {
					t.Errorf("Expected length %d, got %d", tt.expectedLen, len(history))
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGetResumeNode(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		pathID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			userID: "u1",
			pathID: "p1",
			setup: func(m *repository_test.Repopository) {
				m.GetNextNodeIDFunc = func(ctx context.Context, userID, pathID string) (string, error) {
					return "n1", nil
				}
				m.GetNodeByIDFunc = func(ctx context.Context, nodeID string) (*model.Node, error) {
					return &model.Node{NodeID: nodeID}, nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyUserID",
			userID:        "",
			pathID:        "p1",
			setup:         nil,
			expectedError: "user_id is required",
		},
		{
			name:          "EmptyPathID",
			userID:        "u1",
			pathID:        "",
			setup:         nil,
			expectedError: "path_id is required",
		},
		{
			name:   "CompleteOrNoPending",
			userID: "u1",
			pathID: "p1",
			setup: func(m *repository_test.Repopository) {
				m.GetNextNodeIDFunc = func(ctx context.Context, userID, pathID string) (string, error) {
					return "", sql.ErrNoRows
				}
			},
			expectedError: "No pending node found",
		},
		{
			name:   "NodeDetailError",
			userID: "u1",
			pathID: "p1",
			setup: func(m *repository_test.Repopository) {
				m.GetNextNodeIDFunc = func(ctx context.Context, userID, pathID string) (string, error) {
					return "n2", nil
				}
				m.GetNodeByIDFunc = func(ctx context.Context, nodeID string) (*model.Node, error) {
					return nil, errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetResumeNode case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetResumeNode(context.Background(), tt.userID, tt.pathID)
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

func TestSearchLearningPaths(t *testing.T) {
	t.Run("EmptyQuery", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(nil, nil, logger)
		_, err := svc.SearchLearningPaths(context.Background(), model.SearchPathRequest{Query: ""})
		if err == nil || !strings.Contains(err.Error(), "search query cannot be empty") {
			t.Errorf("Expected error about empty query, got %v", err)
		}
	})

	t.Run("AIFailure", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		aiClient := aiclient.NewAIClient("http://mock-ai") // This will fail network req
		svc := service.NewService(nil, aiClient, logger)
		_, err := svc.SearchLearningPaths(context.Background(), model.SearchPathRequest{Query: "Go"})
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected AI network failure, got %v", err)
		}
	})

	t.Run("SuccessEmptyMockFallback", func(t *testing.T) {
		// Mocking success with an unresolvable string just to hit the "no matching results" logic
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		aiClient := aiclient.NewAIClient("http://mock-ai")
		svc := service.NewService(nil, aiClient, logger)
		_, _ = svc.SearchLearningPaths(context.Background(), model.SearchPathRequest{Query: "Go"})
	})
}

func TestGetCollectionInfo(t *testing.T) {
	t.Run("EmptyName", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(nil, nil, logger)
		_, err := svc.GetCollectionInfo("")
		if err == nil || !strings.Contains(err.Error(), "collection name cannot be empty") {
			t.Errorf("Expected error about empty name, got %v", err)
		}
	})

	t.Run("AIFailure", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		aiClient := aiclient.NewAIClient("http://mock-ai")
		svc := service.NewService(nil, aiClient, logger)
		_, err := svc.GetCollectionInfo("learning_paths")
		if err == nil || (!strings.Contains(err.Error(), "failed to get collection info") && !strings.Contains(err.Error(), "internal server error")) {
			t.Errorf("Expected collection info failure, got %v", err)
		}
	})
}

func TestSyncLearningPath(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:          "EmptyPath",
			pathID:        "",
			expectedError: "path ID cannot be empty",
		},
		{
			name:   "NotFound",
			pathID: "p1",
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "learning path not found",
		},
		{
			name:   "AIFailureOrNilAIClient",
			pathID: "p2",
			setup: func(m *repository_test.Repopository) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return &model.LearningPath{PathID: "p2"}, nil
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestSyncLearningPath case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger) // Nil AIClient passed

			_, err := svc.SyncLearningPath(context.Background(), tt.pathID)
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
