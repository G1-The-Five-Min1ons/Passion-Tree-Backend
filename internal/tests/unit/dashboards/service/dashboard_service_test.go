package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/dashboards/model"
	"passiontree/internal/dashboards/service"
	repository_test "passiontree/internal/tests/unit/dashboards/repository"
)

func TestGetDashboardData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetUserInfoFunc: func(ctx context.Context, userID string) (*model.UserInfo, error) {
				return &model.UserInfo{Username: "testuser", FirstName: "Test", Level: 5, XP: 100}, nil
			},
			GetWeeklyMissionsFunc: func(ctx context.Context, userID string) ([]model.MissionItem, error) {
				return []model.MissionItem{{MissionID: "m1", Detail: "Complete a path", RewardXP: 50, Status: "in_progress"}}, nil
			},
			GetCurrentPathsFunc: func(ctx context.Context, userID string) ([]model.CurrentPathItem, error) {
				return []model.CurrentPathItem{{PathID: "p1", Title: "Go Basics", ProgressPercent: 50.0}}, nil
			},
			GetUserActivityFunc: func(ctx context.Context, userID string) ([]model.ActivityItem, error) {
				return []model.ActivityItem{{ActivityType: "enroll", Title: "Enrolled in Go Basics"}}, nil
			},
			GetActivityHeatmapFunc: func(ctx context.Context, userID string) ([]model.ActivityHeatmap, error) {
				return []model.ActivityHeatmap{{Date: "2026-03-01", Count: 3}}, nil
			},
			GetTreeCounterStatsFunc: func(ctx context.Context, userID string) (*model.TreeCounterStats, error) {
				return &model.TreeCounterStats{TotalTreesPlanted: 2, TotalNodesUnlocked: 10}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		resp, err := svc.GetDashboardData(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		if resp.UserInfo.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", resp.UserInfo.Username)
		}
		if len(resp.WeeklyMissions) != 1 {
			t.Errorf("Expected 1 mission, got %d", len(resp.WeeklyMissions))
		}
		if len(resp.CurrentPaths) != 1 {
			t.Errorf("Expected 1 path, got %d", len(resp.CurrentPaths))
		}
		if resp.TreeCounter.TotalTreesPlanted != 2 {
			t.Errorf("Expected 2 trees, got %d", resp.TreeCounter.TotalTreesPlanted)
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, logger)

		_, err := svc.GetDashboardData(context.Background(), "")
		if err == nil || !strings.Contains(err.Error(), "user_id is required") {
			t.Errorf("Expected user_id validation error, got %v", err)
		}
	})

	t.Run("UserInfoError_ReturnsError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetUserInfoFunc: func(ctx context.Context, userID string) (*model.UserInfo, error) {
				return nil, errors.New("db connection failed")
			},
			GetWeeklyMissionsFunc: func(ctx context.Context, userID string) ([]model.MissionItem, error) {
				return []model.MissionItem{}, nil
			},
			GetCurrentPathsFunc: func(ctx context.Context, userID string) ([]model.CurrentPathItem, error) {
				return []model.CurrentPathItem{}, nil
			},
			GetUserActivityFunc: func(ctx context.Context, userID string) ([]model.ActivityItem, error) {
				return []model.ActivityItem{}, nil
			},
			GetActivityHeatmapFunc: func(ctx context.Context, userID string) ([]model.ActivityHeatmap, error) {
				return []model.ActivityHeatmap{}, nil
			},
			GetTreeCounterStatsFunc: func(ctx context.Context, userID string) (*model.TreeCounterStats, error) {
				return &model.TreeCounterStats{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		_, err := svc.GetDashboardData(context.Background(), "user-1")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error for user info failure, got %v", err)
		}
	})

	t.Run("PartialFailure_NonCriticalFallbackToEmpty", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetUserInfoFunc: func(ctx context.Context, userID string) (*model.UserInfo, error) {
				return &model.UserInfo{Username: "testuser"}, nil
			},
			GetWeeklyMissionsFunc: func(ctx context.Context, userID string) ([]model.MissionItem, error) {
				return nil, errors.New("missions db error")
			},
			GetCurrentPathsFunc: func(ctx context.Context, userID string) ([]model.CurrentPathItem, error) {
				return nil, errors.New("paths db error")
			},
			GetUserActivityFunc: func(ctx context.Context, userID string) ([]model.ActivityItem, error) {
				return nil, errors.New("activity db error")
			},
			GetActivityHeatmapFunc: func(ctx context.Context, userID string) ([]model.ActivityHeatmap, error) {
				return nil, errors.New("heatmap db error")
			},
			GetTreeCounterStatsFunc: func(ctx context.Context, userID string) (*model.TreeCounterStats, error) {
				return nil, errors.New("tree stats db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, logger)

		resp, err := svc.GetDashboardData(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("Expected no error (non-critical failures should fallback), got %v", err)
		}
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		// Non-critical fields should fallback to empty
		if len(resp.WeeklyMissions) != 0 {
			t.Errorf("Expected 0 missions (fallback), got %d", len(resp.WeeklyMissions))
		}
		if len(resp.CurrentPaths) != 0 {
			t.Errorf("Expected 0 paths (fallback), got %d", len(resp.CurrentPaths))
		}
	})
}
