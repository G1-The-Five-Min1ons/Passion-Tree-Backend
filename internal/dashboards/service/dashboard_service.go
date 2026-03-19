package service

import (
	"context"
	"sync"

	"passiontree/internal/dashboards/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) GetDashboardData(ctx context.Context, userID string) (*model.DashboardResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	var (
		userInfo   *model.UserInfo
		infoErr    error 
		missions   []model.MissionItem
		paths      []model.CurrentPathItem
		activities []model.ActivityItem
		heatmap    []model.ActivityHeatmap
		treeStats  *model.TreeCounterStats

		wg sync.WaitGroup // ใช้ WaitGroup เพื่อรอให้ทุก Goroutines ทำงานเสร็จ
	)

	wg.Add(6)

	// 1. User Info
	go func() {
		defer wg.Done()
		userInfo, infoErr = s.dashboardrepo.GetUserInfo(ctx, userID)
		if infoErr != nil {
			s.logger.ErrorContext(ctx, "failed to fetch user info", "error", infoErr, "user_id", userID)
		}
	}()

	// 2. Weekly Missions
	go func() {
		defer wg.Done()
		var err error
		missions, err = s.dashboardrepo.GetWeeklyMissions(ctx, userID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to fetch missions", "error", err, "user_id", userID)
			missions = []model.MissionItem{} // Fallback to empty
		}
	}()

	// 3. Current Paths
	go func() {
		defer wg.Done()
		var err error
		paths, err = s.dashboardrepo.GetCurrentPaths(ctx, userID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to fetch current paths", "error", err, "user_id", userID)
			paths = []model.CurrentPathItem{}
		}
	}()

	// 4. Recent Activity
	go func() {
		defer wg.Done()
		var err error
		activities, err = s.dashboardrepo.GetUserActivity(ctx, userID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to fetch activities", "error", err, "user_id", userID)
			activities = []model.ActivityItem{}
		}
	}()

	// 5. Activity Heatmap
	go func() {
		defer wg.Done()
		var err error
		heatmap, err = s.dashboardrepo.GetActivityHeatmap(ctx, userID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to fetch heatmap", "error", err, "user_id", userID)
			heatmap = []model.ActivityHeatmap{}
		}
	}()

	// 6. Fetch Tree Stats
	go func() {
		defer wg.Done()
		var err error
		treeStats, err = s.dashboardrepo.GetTreeCounterStats(ctx, userID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to fetch tree stats", "error", err, "user_id", userID)
			treeStats = &model.TreeCounterStats{TotalTreesPlanted: 0, TotalNodesUnlocked: 0}
		}
	}()

	wg.Wait()

	if infoErr != nil {
		return nil, apperror.NewInternal("failed to fetch user information")
	}

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
