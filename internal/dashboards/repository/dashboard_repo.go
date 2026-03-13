package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/dashboards/model"
)

func (r *repositoryImpl) GetUserInfo(ctx context.Context, userID string) (*model.UserInfo, error) {
	query := `
		SELECT 
			u.username, ISNULL(u.first_name, ''), ISNULL(p.avatar_url, ''), 
			ISNULL(p.level, 1), ISNULL(p.xp, 0), ISNULL(p.rank_name, 'Beginner'), 
			ISNULL(p.learning_streak, 0), ISNULL(p.hour_learned, 0)
		FROM Users u
		LEFT JOIN Profile p ON u.user_id = p.user_id
		WHERE u.user_id = @p1
	`
	var info model.UserInfo
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&info.Username, &info.FirstName, &info.AvatarURL,
		&info.Level, &info.XP, &info.RankName, &info.LearningStreak, &info.HourLearned,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("repo.GetUserInfo query failed: %w", err)
	}
	return &info, nil
}

func (r *repositoryImpl) GetWeeklyMissions(ctx context.Context, userID string) ([]model.MissionItem, error) {
	// ดึง Mission ที่ยังไม่หมดอายุ หรือเพิ่งทำเสร็จในสัปดาห์นี้
	query := `
		SELECT 
			CONVERT(VARCHAR(36), m.mission_id), m.detail, ISNULL(m.reward_xp, 0), ISNULL(um.status, 'pending'), um.expire_at
		FROM User_Mission um
		JOIN Mission m ON um.mission_id = m.mission_id
		WHERE um.user_id = @p1 AND um.expire_at >= GETDATE()
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetWeeklyMissions query failed: %w", err)
	}
	defer rows.Close()

	var missions []model.MissionItem
	for rows.Next() {
		var m model.MissionItem
		if err := rows.Scan(&m.MissionID, &m.Detail, &m.RewardXP, &m.Status, &m.ExpireAt); err != nil {
			return nil, fmt.Errorf("repo.GetWeeklyMissions scan failed: %w", err)
		}
		missions = append(missions, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetWeeklyMissions row iteration failed: %w", err)
	}

	return missions, nil
}

func (r *repositoryImpl) GetCurrentPaths(ctx context.Context, userID string) ([]model.CurrentPathItem, error) {
	// คำนวณ Progress แบบคร่าวๆ (Nodes completed / Total Nodes)
	query := `
		SELECT 
			CONVERT(VARCHAR(36), lp.path_id), lp.title, ISNULL(lp.cover_img_url, ''),
			CASE 
				WHEN ISNULL(n_count.total_nodes, 0) = 0 THEN 0.0
				ELSE ROUND((CAST(ISNULL(progress_count.completed, 0) AS FLOAT) / CAST(n_count.total_nodes AS FLOAT)) * 100, 2)
			END as progress_percent
		FROM Path_Enroll pe
		JOIN Learning_Path lp ON pe.path_id = lp.path_id
		LEFT JOIN (SELECT path_id, COUNT(node_id) as total_nodes FROM Node GROUP BY path_id) AS n_count ON lp.path_id = n_count.path_id
		LEFT JOIN (
			SELECT n.path_id, COUNT(DISTINCT np.node_id) as completed
			FROM Node_progress np JOIN Node n ON np.node_id = n.node_id
			WHERE np.user_id = @p1 AND np.complete = 'true'
			GROUP BY n.path_id
		) AS progress_count ON lp.path_id = progress_count.path_id
		WHERE pe.user_id = @p1 AND pe.enrollment_status = 'active'
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetCurrentPaths query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.CurrentPathItem
	for rows.Next() {
		var p model.CurrentPathItem
		if err := rows.Scan(&p.PathID, &p.Title, &p.CoverImgURL, &p.ProgressPercent); err != nil {
			return nil, fmt.Errorf("repo.QueryContext scan failed: %w", err)
		}
		paths = append(paths, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.QueryContext row iteration failed: %w", err)
	}

	return paths, nil
}

func (r *repositoryImpl) GetUserActivity(ctx context.Context, userID string) ([]model.ActivityItem, error) {
	// UNION ข้อมูล Activity ต่างๆ เข้าด้วยกัน (Enroll, Complete Node, Complete Mission)
	query := `
		SELECT TOP 20 activity_type, title, timestamp FROM (
			SELECT 'enroll_path' as activity_type, lp.title as title, pe.enroll_at as timestamp
			FROM Path_Enroll pe JOIN Learning_Path lp ON pe.path_id = lp.path_id WHERE pe.user_id = @p1
			UNION ALL
			SELECT 'complete_node', n.title, np.updated_at
			FROM Node_progress np JOIN Node n ON np.node_id = n.node_id WHERE np.user_id = @p1 AND np.complete = 'true'
			UNION ALL
			SELECT 'complete_mission', m.detail, um.complete_at
			FROM User_Mission um JOIN Mission m ON um.mission_id = m.mission_id WHERE um.user_id = @p1 AND um.status = 'completed'
		) as combined_activities
		ORDER BY timestamp DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetUserActivity query failed: %w", err)
	}
	defer rows.Close()

	var activities []model.ActivityItem
	for rows.Next() {
		var a model.ActivityItem
		if err := rows.Scan(&a.ActivityType, &a.Title, &a.Timestamp); err != nil {
			return nil, fmt.Errorf("repo.GetUserActivity scan failed: %w", err)
		}
		activities = append(activities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetUserActivity row iteration failed: %w", err)
	}

	return activities, nil
}

func (r *repositoryImpl) GetActivityHeatmap(ctx context.Context, userID string) ([]model.ActivityHeatmap, error) {
	// นับจำนวน Activity รายวันใน 365 วันย้อนหลัง สำหรับ GitHub Graph
	query := `
		SELECT FORMAT(timestamp, 'yyyy-MM-dd') as date, COUNT(*) as count
		FROM (
			SELECT pe.enroll_at as timestamp FROM Path_Enroll pe WHERE pe.user_id = @p1
			UNION ALL
			SELECT np.updated_at FROM Node_progress np WHERE np.user_id = @p1 AND np.complete = 'true'
			UNION ALL
			SELECT um.complete_at FROM User_Mission um WHERE um.user_id = @p1 AND um.status = 'completed'
		) as combined
		WHERE timestamp >= DATEADD(day, -365, GETDATE())
		GROUP BY FORMAT(timestamp, 'yyyy-MM-dd')
		ORDER BY date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetActivityHeatmap query failed: %w", err)
	}
	defer rows.Close()

	var heatmap []model.ActivityHeatmap
	for rows.Next() {
		var h model.ActivityHeatmap
		if err := rows.Scan(&h.Date, &h.Count); err != nil {
			return nil, fmt.Errorf("repo.GetActivityHeatmap scan failed: %w", err)
		}
		heatmap = append(heatmap, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetActivityHeatmap row iteration failed: %w", err)
	}

	return heatmap, nil
}

func (r *repositoryImpl) GetTreeCounterStats(ctx context.Context, userID string) (*model.TreeCounterStats, error) {
	query := `
		SELECT 
			COUNT(t.tree_id) AS total_trees_planted,
			ISNULL(SUM(t.node_count), 0) AS total_nodes_unlocked
		FROM Tree t
		JOIN Tree_Album ta ON t.album_id = ta.album_id
		WHERE ta.user_id = @p1
	`

	var stats model.TreeCounterStats
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalTreesPlanted,
		&stats.TotalNodesUnlocked,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &model.TreeCounterStats{TotalTreesPlanted: 0, TotalNodesUnlocked: 0}, nil
		}
		return nil, fmt.Errorf("repo.GetTreeCounterStats query failed: %w", err)
	}

	return &stats, nil
}
