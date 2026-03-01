package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"
	"passiontree/internal/recommendation/service"
	repository_test "passiontree/internal/tests/unit/recommendation/repository"
)

func TestRecommendPathsForUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		treeID        string
		setup         func(*repository_test.Repopository)
		expectedError string
		expectedPaths int
	}{
		{
			name:          "MissingParams_EmptyUserID",
			userID:        "",
			treeID:        "tree-123",
			setup:         nil,
			expectedError: "user_id and tree_id are required",
		},
		{
			name:          "MissingParams_EmptyTreeID",
			userID:        "user-123",
			treeID:        "",
			setup:         nil,
			expectedError: "user_id and tree_id are required",
		},
		{
			name:   "RepoError_ReturnsInternalError",
			userID: "user-123",
			treeID: "tree-123",
			setup: func(m *repository_test.Repopository) {
				m.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return nil, "", errors.New("database connection failed")
				}
			},
			expectedError: "internal server error",
		},
		{
			name:   "Success_NoReflections_ReturnsEmptyDataEarly",
			userID: "user-123",
			treeID: "tree-123",
			setup: func(m *repository_test.Repopository) {
				m.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return []model.UserReflection{}, "", nil
				}
			},
			expectedError: "",
			expectedPaths: 0,
		},
		{
			name:   "AIFailure_WhenCallingAIClient",
			userID: "user-123",
			treeID: "tree-123",
			setup: func(m *repository_test.Repopository) {
				m.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					mockReflections := []model.UserReflection{
						{Summary: "Learned variables", PrimaryEmotion: "Happy"},
					}
					return mockReflections, "path-1", nil
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRecommendPathsForUser case: %s\033[0m", tt.name)

			mockRepo := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			aiClient := aiclient.NewAIClient("http://mock-ai-endpoint")

			svc := service.NewService(mockRepo, aiClient, logger)

			resp, err := svc.RecommendPathsForUser(context.Background(), tt.userID, tt.treeID)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if resp != nil && len(resp.RecommendedPaths) != tt.expectedPaths {
					t.Errorf("Expected %d paths, got %d", tt.expectedPaths, len(resp.RecommendedPaths))
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}
