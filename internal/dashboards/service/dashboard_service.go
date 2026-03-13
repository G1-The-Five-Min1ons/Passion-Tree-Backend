package service

import (
	"context"
	"passiontree/internal/dashboards/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) GetDashboardData(ctx context.Context, userID string) (*model.DashboardResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	// 1. User Info
	userInfo, err := s.dashboardrepo.GetUserInfo(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch user info", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to fetch user information")
	}

	// 2. Weekly Missions
	missions, err := s.dashboardrepo.GetWeeklyMissions(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch missions", "error", err, "user_id", userID)
		missions = []model.MissionItem{} // Fallback to empty
	}

	// 3. Current Paths
	paths, err := s.dashboardrepo.GetCurrentPaths(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch current paths", "error", err, "user_id", userID)
		paths = []model.CurrentPathItem{}
	}

	// 4. Recent Activity
	activities, err := s.dashboardrepo.GetUserActivity(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch activities", "error", err, "user_id", userID)
		activities = []model.ActivityItem{}
	}

	// 5. Activity Heatmap (GitHub Style)
	heatmap, err := s.dashboardrepo.GetActivityHeatmap(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch heatmap", "error", err, "user_id", userID)
		heatmap = []model.ActivityHeatmap{}
	}

	// 6. Fetch Tree Stats
	treeStats, err := s.dashboardrepo.GetTreeCounterStats(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch tree stats", "error", err, "user_id", userID)
		treeStats = &model.TreeCounterStats{TotalTreesPlanted: 0, TotalNodesUnlocked: 0}
	}

	// ประกอบร่าง Data
	response := &model.DashboardResponse{
		UserInfo:        *userInfo,
		WeeklyMissions:  missions,
		CurrentPaths:    paths,
		RecentActivity:  activities,
		ActivitySummary: heatmap,
		TreeCounter:     *treeStats,
	}

	s.logger.InfoContext(ctx, "dashboard data aggregated successfully", "user_id", userID)
	return response, nil
}
