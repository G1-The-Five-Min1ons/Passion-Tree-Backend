package repository

import (
	"context"

	"passiontree/internal/dashboards/model"
)

// MockRepo for Dashboards
type Repository struct {
	GetUserInfoFunc         func(ctx context.Context, userID string) (*model.UserInfo, error)
	GetWeeklyMissionsFunc   func(ctx context.Context, userID string) ([]model.MissionItem, error)
	GetCurrentPathsFunc     func(ctx context.Context, userID string) ([]model.CurrentPathItem, error)
	GetUserActivityFunc     func(ctx context.Context, userID string) ([]model.ActivityItem, error)
	GetActivityHeatmapFunc  func(ctx context.Context, userID string) ([]model.ActivityHeatmap, error)
	GetTreeCounterStatsFunc func(ctx context.Context, userID string) (*model.TreeCounterStats, error)
}

func (m *Repository) GetUserInfo(ctx context.Context, userID string) (*model.UserInfo, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(ctx, userID)
	}
	return nil, nil
}

func (m *Repository) GetWeeklyMissions(ctx context.Context, userID string) ([]model.MissionItem, error) {
	if m.GetWeeklyMissionsFunc != nil {
		return m.GetWeeklyMissionsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *Repository) GetCurrentPaths(ctx context.Context, userID string) ([]model.CurrentPathItem, error) {
	if m.GetCurrentPathsFunc != nil {
		return m.GetCurrentPathsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *Repository) GetUserActivity(ctx context.Context, userID string) ([]model.ActivityItem, error) {
	if m.GetUserActivityFunc != nil {
		return m.GetUserActivityFunc(ctx, userID)
	}
	return nil, nil
}

func (m *Repository) GetActivityHeatmap(ctx context.Context, userID string) ([]model.ActivityHeatmap, error) {
	if m.GetActivityHeatmapFunc != nil {
		return m.GetActivityHeatmapFunc(ctx, userID)
	}
	return nil, nil
}

func (m *Repository) GetTreeCounterStats(ctx context.Context, userID string) (*model.TreeCounterStats, error) {
	if m.GetTreeCounterStatsFunc != nil {
		return m.GetTreeCounterStatsFunc(ctx, userID)
	}
	return nil, nil
}
