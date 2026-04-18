package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	lpmodel "passiontree/internal/learning-path/model"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"
	"passiontree/internal/recommendation/service"
	repository_test "passiontree/internal/tests/unit/recommendation/repository"
)

func setupMockUIServer(responseBody string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

func TestRecommendPathsForUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		treeID        string
		aiResponse    string
		aiStatus      int
		setup         func(*repository_test.MockRecRepository, *repository_test.MockPathRepository)
		expectedError string
		expectedPaths int
	}{
		{
			name:          "MissingParams_EmptyUserID",
			userID:        "",
			treeID:        "tree-123",
			expectedError: "Authentication session expired",
		},
		{
			name:   "RepoError_ReturnsInternalError",
			userID: "user-123",
			treeID: "tree-123",
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return nil, "", errors.New("database connection failed")
				}
			},
			expectedError: "internal server error",
		},
		{
			name:   "Success_NoReflections_ReturnsEmptyDataEarly",
			userID: "user-123",
			treeID: "tree-123",
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return []model.UserReflection{}, "", nil
				}
			},
			expectedPaths: 0,
		},
		{
			name:       "Success_WithAIResults",
			userID:     "user-123",
			treeID:     "tree-123",
			aiResponse: `{"results": [{"id": "path-2", "score": 0.95, "payload": {"title": "Advanced Go"}}, {"id": "path-1", "score": 0.85}]}`,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return []model.UserReflection{
						{Summary: "Learned APIs", PrimaryEmotion: "Excited"},
					}, "path-1", nil
				}
				path.GetLearningPathsByIDsFunc = func(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error) {
					return []lpmodel.LearningPath{
						{PathID: "PATH-2", Title: "Advanced Go"},
					}, nil
				}
			},
			expectedPaths: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRecommendPathsForUser case: %s\033[0m", tt.name)

			recRepo := &repository_test.MockRecRepository{}
			pathRepo := &repository_test.MockPathRepository{}
			if tt.setup != nil {
				tt.setup(recRepo, pathRepo)
			}

			ts := setupMockUIServer(tt.aiResponse, tt.aiStatus)
			defer ts.Close()

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			aiClient := aiclient.NewAIClient(ts.URL)

			svc := service.NewService(recRepo, pathRepo, aiClient, logger)

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

func TestRecommendHomePathsForUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		aiResponse    string
		aiStatus      int
		setup         func(*repository_test.MockRecRepository, *repository_test.MockPathRepository)
		expectedError string
		expectedPaths int
	}{
		{
			name:          "MissingParams_EmptyUserID",
			userID:        "",
			expectedError: "Authentication session expired",
		},
		{
			name:   "NoEnrolledPaths_ReturnsTopPopular",
			userID: "user-new",
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserEnrolledPathsForRecFunc = func(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{}, nil
				}
				rec.GetTopPopularPathsFunc = func(ctx context.Context) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{{LearningPath: lpmodel.LearningPath{PathID: "pop-1"}}, {LearningPath: lpmodel.LearningPath{PathID: "pop-2"}}}, nil
				}
			},
			expectedPaths: 2,
		},
		{
			name:       "Success_WithEnrolledPaths_FilteredAIResults",
			userID:     "user-123",
			aiResponse: `{"results": [{"id": "path-new", "score": 0.9}, {"id": "path-enrolled", "score": 0.8}]}`,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserEnrolledPathsForRecFunc = func(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{{LearningPath: lpmodel.LearningPath{PathID: "path-enrolled", Title: "Basic Go"}}}, nil
				}
				path.GetLearningPathsByIDsFunc = func(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error) {
					return []lpmodel.LearningPath{
						{PathID: "PATH-NEW", Title: "Found Path"},
					}, nil
				}
			},
			expectedPaths: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRecommendHomePathsForUser case: %s\033[0m", tt.name)

			recRepo := &repository_test.MockRecRepository{}
			pathRepo := &repository_test.MockPathRepository{}
			if tt.setup != nil {
				tt.setup(recRepo, pathRepo)
			}

			ts := setupMockUIServer(tt.aiResponse, tt.aiStatus)
			defer ts.Close()

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			aiClient := aiclient.NewAIClient(ts.URL)

			svc := service.NewService(recRepo, pathRepo, aiClient, logger)

			resp, err := svc.RecommendHomePathsForUser(context.Background(), tt.userID)

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
