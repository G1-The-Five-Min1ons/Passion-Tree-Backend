package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"
	"passiontree/internal/recommendation/service"
	repository_test "passiontree/internal/tests/unit/recommendation/repository"
)

func TestRunDailyRecommendationBatch(t *testing.T) {
	sampleInteractions := []model.UserInteraction{
		{UserID: "u1", PathID: "p1", Score: 3.0},
	}
	sampleProfiles := []model.UserProfile{
		{UserID: "u1", Interests: "Interested in: golang. Learning style: visual."},
	}
	batchAIResponse := `{
		"success": true,
		"message": "ok",
		"data": [{"user_id": "u1", "recommended_paths": [{"path_id": "p2", "score": 0.91}]}]
	}`

	tests := []struct {
		name          string
		nilAIClient   bool
		aiResponse    string
		aiStatus      int
		setup         func(*repository_test.MockRecRepository)
		expectedError string
	}{
		{
			name:          "NilAIClient_AbortsBatch",
			nilAIClient:   true,
			expectedError: "ai client is not initialized",
		},
		{
			name:     "GetBatchInteractions_Error",
			aiStatus: http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository) {
				rec.GetBatchInteractionsFunc = func(ctx context.Context) ([]model.UserInteraction, error) {
					return nil, errors.New("interactions table locked")
				}
			},
			expectedError: "interactions table locked",
		},
		{
			name:     "GetBatchProfiles_Error",
			aiStatus: http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository) {
				rec.GetBatchInteractionsFunc = func(ctx context.Context) ([]model.UserInteraction, error) {
					return sampleInteractions, nil
				}
				rec.GetBatchProfilesFunc = func(ctx context.Context) ([]model.UserProfile, error) {
					return nil, errors.New("profiles table unavailable")
				}
			},
			expectedError: "profiles table unavailable",
		},
		{
			name:       "AIComputationFailed_ReturnsError",
			aiResponse: `{"error": "model overloaded"}`,
			aiStatus:   http.StatusServiceUnavailable,
			setup: func(rec *repository_test.MockRecRepository) {
				rec.GetBatchInteractionsFunc = func(ctx context.Context) ([]model.UserInteraction, error) {
					return sampleInteractions, nil
				}
				rec.GetBatchProfilesFunc = func(ctx context.Context) ([]model.UserProfile, error) {
					return sampleProfiles, nil
				}
			},
			expectedError: "AI service returned status",
		},
		{
			// Simulates the reconcile step: delete-then-insert per user fails midway.
			name:       "ReconcilePartialFailure_SaveError",
			aiResponse: batchAIResponse,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository) {
				rec.GetBatchInteractionsFunc = func(ctx context.Context) ([]model.UserInteraction, error) {
					return sampleInteractions, nil
				}
				rec.GetBatchProfilesFunc = func(ctx context.Context) ([]model.UserProfile, error) {
					return sampleProfiles, nil
				}
				rec.SaveBatchRecommendationsFunc = func(ctx context.Context, results []model.BatchRecommendationResult) error {
					return errors.New("failed to delete old recommendations for user u1: deadlock detected")
				}
			},
			expectedError: "failed to delete old recommendations for user u1",
		},
		{
			name:       "Success_BatchCompleted",
			aiResponse: batchAIResponse,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository) {
				rec.GetBatchInteractionsFunc = func(ctx context.Context) ([]model.UserInteraction, error) {
					return sampleInteractions, nil
				}
				rec.GetBatchProfilesFunc = func(ctx context.Context) ([]model.UserProfile, error) {
					return sampleProfiles, nil
				}
				rec.SaveBatchRecommendationsFunc = func(ctx context.Context, results []model.BatchRecommendationResult) error {
					return nil
				}
			},
		},
		{
			name:       "EmptyPayload_StillSaves",
			aiResponse: `{"success": true, "message": "ok", "data": []}`,
			aiStatus:   http.StatusOK,
			setup: func(rec *repository_test.MockRecRepository) {
				rec.GetBatchInteractionsFunc = func(ctx context.Context) ([]model.UserInteraction, error) {
					return []model.UserInteraction{}, nil
				}
				rec.GetBatchProfilesFunc = func(ctx context.Context) ([]model.UserProfile, error) {
					return []model.UserProfile{}, nil
				}
				rec.SaveBatchRecommendationsFunc = func(ctx context.Context, results []model.BatchRecommendationResult) error {
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRunDailyRecommendationBatch case: %s\033[0m", tt.name)

			recRepo := &repository_test.MockRecRepository{}
			if tt.setup != nil {
				tt.setup(recRepo)
			}
			pathRepo := &repository_test.MockPathRepository{}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			var ai *aiclient.AIClient
			if !tt.nilAIClient {
				aiResp := tt.aiResponse
				if aiResp == "" {
					aiResp = `{"success": true, "message": "ok", "data": []}`
				}
				status := tt.aiStatus
				if status == 0 {
					status = http.StatusOK
				}
				ts := setupMockUIServer(aiResp, status)
				defer ts.Close()
				ai = aiclient.NewAIClient(ts.URL)
			}

			svc := service.NewService(recRepo, pathRepo, ai, logger)
			err := svc.RunDailyRecommendationBatch(context.Background())

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %v", tt.expectedError, err)
				}
			}
		})
	}
}
