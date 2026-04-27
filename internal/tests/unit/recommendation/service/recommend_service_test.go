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
			name:          "EmptyUserID_ReturnsUnauthorized",
			userID:        "",
			treeID:        "tree-123",
			expectedError: "Authentication session expired",
		},
		{
			name:          "EmptyTreeID_ReturnsBadRequest",
			userID:        "user-123",
			treeID:        "",
			expectedError: "please select a specific tree to get recommendations",
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
			name:   "NoReflections_ReturnsEmptyEarly",
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
			name:       "AIResult_WithFullPayload_OverridesDBFields",
			userID:     "user-123",
			treeID:     "tree-123",
			aiResponse: `{"results": [{"id": "path-2", "score": 0.95, "payload": {"title": "Advanced Go", "cover_img_url": "https://img.example.com/go.png", "objective": "Master Go"}}, {"id": "path-1", "score": 0.85}]}`,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return []model.UserReflection{
						{Summary: "Learned APIs", PrimaryEmotion: "Excited"},
					}, "path-1", nil
				}
				path.GetLearningPathsByIDsFunc = func(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error) {
					var result []lpmodel.LearningPath
					for _, id := range pathIDs {
						if id == "path-2" {
							result = append(result, lpmodel.LearningPath{PathID: "path-2", Title: "DB Title"})
						}
					}
					return result, nil
				}
			},
			expectedPaths: 1,
		},
		{
			name:       "AIResult_NilPayload_UsesDBFields",
			userID:     "user-123",
			treeID:     "tree-123",
			aiResponse: `{"results": [{"id": "path-2", "score": 0.95}, {"id": "path-1", "score": 0.85}]}`,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return []model.UserReflection{
						{Summary: "Learned concurrency", PrimaryEmotion: "Confident"},
					}, "path-1", nil
				}
				path.GetLearningPathsByIDsFunc = func(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error) {
					return []lpmodel.LearningPath{
						{PathID: "path-2", Title: "Go Concurrency", Objective: "Understand goroutines"},
					}, nil
				}
			},
			expectedPaths: 1,
		},
		{
			name:       "AIResult_PartialPayload_MergesWithDB",
			userID:     "user-123",
			treeID:     "tree-123",
			aiResponse: `{"results": [{"id": "path-2", "score": 0.90, "payload": {"title": "AI Title"}}]}`,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return []model.UserReflection{
						{Summary: "Channels deep dive", PrimaryEmotion: "Curious"},
					}, "path-1", nil
				}
				path.GetLearningPathsByIDsFunc = func(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error) {
					return []lpmodel.LearningPath{
						{PathID: "path-2", Title: "DB Title", CoverImgURL: "https://db.example.com/img.png"},
					}, nil
				}
			},
			expectedPaths: 1,
		},
		{
			name:       "PathRepoError_FallsBackToPopular",
			userID:     "user-123",
			treeID:     "tree-123",
			aiResponse: `{"results": [{"id": "path-2", "score": 0.9}]}`,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetUserReflectionsByTreeFunc = func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
					return []model.UserReflection{
						{Summary: "Test topic", PrimaryEmotion: "Motivated"},
					}, "path-1", nil
				}
				path.GetLearningPathsByIDsFunc = func(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error) {
					return nil, errors.New("DB unavailable")
				}
				rec.GetTopPopularPathsFunc = func(ctx context.Context) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{
						{LearningPath: lpmodel.LearningPath{PathID: "popular-1"}},
						{LearningPath: lpmodel.LearningPath{PathID: "popular-2"}},
					}, nil
				}
			},
			expectedPaths: 2,
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

			aiResponse := tt.aiResponse
			if aiResponse == "" {
				aiResponse = `{"results": []}`
			}
			aiStatus := tt.aiStatus
			if aiStatus == 0 {
				aiStatus = http.StatusOK
			}

			ts := setupMockUIServer(aiResponse, aiStatus)
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
					t.Errorf("Expected error containing %q, got %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestRecommendHomePathsForUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setup         func(*repository_test.MockRecRepository, *repository_test.MockPathRepository)
		expectedError string
		expectedPaths int
	}{
		{
			name:          "EmptyUserID_ReturnsUnauthorized",
			userID:        "",
			expectedError: "Authentication session expired",
		},
		{
			name:   "RepoError_ReturnsInternalError",
			userID: "user-123",
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetSavedHomeRecommendationsFunc = func(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
					return nil, errors.New("db timeout")
				}
			},
			expectedError: "internal server error",
		},
		{
			name:   "NoSavedRecommendations_FallsBackToPopular",
			userID: "user-new",
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetSavedHomeRecommendationsFunc = func(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{}, nil
				}
				rec.GetTopPopularPathsFunc = func(ctx context.Context) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{
						{LearningPath: lpmodel.LearningPath{PathID: "pop-1"}},
						{LearningPath: lpmodel.LearningPath{PathID: "pop-2"}},
					}, nil
				}
			},
			expectedPaths: 2,
		},
		{
			name:   "HasSavedRecommendations_ReturnsThem",
			userID: "user-123",
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetSavedHomeRecommendationsFunc = func(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{
						{LearningPath: lpmodel.LearningPath{PathID: "saved-1"}},
						{LearningPath: lpmodel.LearningPath{PathID: "saved-2"}},
						{LearningPath: lpmodel.LearningPath{PathID: "saved-3"}},
					}, nil
				}
			},
			expectedPaths: 3,
		},
		{
			name:   "FallbackPopularError_ReturnsInternalError",
			userID: "user-no-recs",
			setup: func(rec *repository_test.MockRecRepository, path *repository_test.MockPathRepository) {
				rec.GetSavedHomeRecommendationsFunc = func(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
					return []model.RecommendedPath{}, nil
				}
				rec.GetTopPopularPathsFunc = func(ctx context.Context) ([]model.RecommendedPath, error) {
					return nil, errors.New("popular paths query failed")
				}
			},
			expectedError: "internal server error",
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

			ts := setupMockUIServer(`{"results": []}`, http.StatusOK)
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
					t.Errorf("Expected error containing %q, got %v", tt.expectedError, err)
				}
			}
		})
	}
}
