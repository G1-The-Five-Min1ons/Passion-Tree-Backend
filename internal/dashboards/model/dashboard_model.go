package model

import "time"

type DashboardResponse struct {
	UserInfo        UserInfo          `json:"user_information"`
	WeeklyMissions  []MissionItem     `json:"weekly_missions"`
	CurrentPaths    []CurrentPathItem `json:"current_paths"`
	TreeCounter     TreeCounterStats  `json:"tree_counter"`
	RecentActivity  []ActivityItem    `json:"recent_activity"`
	ActivitySummary []ActivityHeatmap `json:"activity_summary"`
}

type UserInfo struct {
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	AvatarURL      string `json:"avatar_url"`
	Level          int    `json:"level"`
	XP             int64  `json:"xp"`
	RankName       string `json:"rank_name"`
	LearningStreak int    `json:"learning_streak"`
	HourLearned    int    `json:"hour_learned"`
}

type MissionItem struct {
	MissionID string     `json:"mission_id"`
	Detail    string     `json:"detail"`
	RewardXP  int64      `json:"reward_xp"`
	Status    string     `json:"status"`
	ExpireAt  *time.Time `json:"expire_at"`
}

type CurrentPathItem struct {
	PathID          string  `json:"path_id"`
	Title           string  `json:"title"`
	CoverImgURL     string  `json:"cover_img_url"`
	ProgressPercent float64 `json:"progress_percent"`
}

type TreeCounterStats struct {
	TotalTreesPlanted  int `json:"total_trees_planted"`
	TotalNodesUnlocked int `json:"total_nodes_unlocked"`
}

type ActivityItem struct {
	ActivityType string    `json:"activity_type"`
	Title        string    `json:"title"`
	Timestamp    time.Time `json:"timestamp"`
}

type ActivityHeatmap struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
