package repository

import (
	"context"
	"fmt"
	"time"
	"database/sql"

	"github.com/google/uuid"

	"passiontree/internal/mission/model"
)

func (r *repositoryImpl) CreateTemplate(ctx context.Context, req model.CreateTemplateRequest) (string, error) {
	newID := uuid.New().String()
	query := `
		INSERT INTO mission (mission_id, title, detail, reward_xp, reward_heart, condition_type, target_value, is_active, create_at, update_at) 
		OUTPUT INSERTED.mission_id
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, 1, GETDATE(), GETDATE())`

	err := r.db.QueryRowContext(ctx, query, newID, req.Title, req.Description, req.RewardXP, req.RewardHeart, req.ConditionType, req.TargetValue).Scan(&newID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateTemplate failed: %w", err)
	}
	return newID, nil
}

func (r *repositoryImpl) GetActiveTemplates(ctx context.Context) ([]model.MissionTemplate, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), mission_id), title, detail, reward_xp, reward_heart, condition_type, target_value 
		FROM mission WHERE is_active = 1`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repo.GetActiveTemplates failed: %w", err)
	}
	defer rows.Close()

	var templates []model.MissionTemplate
	for rows.Next() {
		var t model.MissionTemplate
		if err := rows.Scan(&t.MissionID, &t.Title, &t.Description, &t.RewardXP, &t.RewardHeart, &t.ConditionType, &t.TargetValue); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetActiveTemplates row iteration failed: %w", err)
	}

	return templates, nil
}

func (r *repositoryImpl) AssignMissionToUser(ctx context.Context, userID string, missionID string, expireAt time.Time) error {
	newID := uuid.New().String()
	query := `
		INSERT INTO user_mission (user_mission_id, user_id, mission_id, current_value, status, expire_at) 
		VALUES (@p1, @p2, @p3, 0, 'active', @p4)`

	_, err := r.db.ExecContext(ctx, query, newID, userID, missionID, expireAt)
	if err != nil {
		return fmt.Errorf("repo.AssignMission failed: %w", err)
	}
	return nil
}

func (r *repositoryImpl) GetUserActiveMissions(ctx context.Context, userID string) ([]model.UserMission, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), um.user_mission_id), um.mission_id, m.title, m.detail, 
			m.reward_xp, m.reward_heart, um.current_value, m.target_value, um.status, um.expire_at
		FROM user_mission um
		JOIN mission m ON um.mission_id = m.mission_id
		WHERE um.user_id = @p1 AND um.status = 'active' AND um.expire_at > GETDATE()`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetUserActiveMissions failed: %w", err)
	}
	defer rows.Close()

	var missions []model.UserMission
	for rows.Next() {
		var m model.UserMission
		if err := rows.Scan(&m.UserMissionID, &m.MissionID, &m.Title, &m.Description, &m.RewardXP, &m.RewardHeart, &m.CurrentValue, &m.TargetValue, &m.Status, &m.ExpireAt); err != nil {
			return nil, err
		}
		m.UserID = userID
		missions = append(missions, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetUserActiveMissions row iteration failed: %w", err)
	}

	return missions, nil
}

func (r *repositoryImpl) GetAllActiveUsers(ctx context.Context) ([]string, error) {
	query := `SELECT CONVERT(VARCHAR(36), user_id) FROM users`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		users = append(users, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetAllActiveUsers row iteration failed: %w", err)
	}

	return users, nil
}

func (r *repositoryImpl) GetActiveMissionsByCondition(ctx context.Context, userID string, conditionType string) ([]model.UserMission, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), um.user_mission_id), um.mission_id, m.title, 
			m.reward_xp, m.reward_heart, um.current_value, m.target_value, um.status
		FROM user_mission um
		JOIN mission m ON um.mission_id = m.mission_id
		WHERE um.user_id = @p1 
		  AND m.condition_type = @p2 
		  AND um.status = 'active' 
		  AND um.expire_at > GETDATE()`

	rows, err := r.db.QueryContext(ctx, query, userID, conditionType)
	if err != nil {
		return nil, fmt.Errorf("repo.GetActiveMissionsByCondition failed: %w", err)
	}
	defer rows.Close()

	var missions []model.UserMission
	for rows.Next() {
		var m model.UserMission
		if err := rows.Scan(&m.UserMissionID, &m.MissionID, &m.Title, &m.RewardXP, &m.RewardHeart, &m.CurrentValue, &m.TargetValue, &m.Status); err != nil {
			return nil, err
		}
		missions = append(missions, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetActiveMissionsByCondition row iteration failed: %w", err)
	}

	return missions, nil
}

func (r *repositoryImpl) UpdateMissionProgressAndReward(ctx context.Context, userMissionID string, userID string, newValue int, isCompleted bool, rewardXP int64, rewardHeart int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	status := "active"
	completeAt := sql.NullTime{Valid: false}
	if isCompleted {
		status = "completed"
		completeAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	updateMissionQ := `
		UPDATE user_mission 
		SET current_value = @p1, status = @p2, complete_at = @p3 
		WHERE user_mission_id = @p4`
	_, err = tx.ExecContext(ctx, updateMissionQ, newValue, status, completeAt, userMissionID)
	if err != nil {
		return fmt.Errorf("failed to update user_mission: %w", err)
	}

	if isCompleted && rewardXP > 0 {
		updateXPQ := `UPDATE profile SET xp = ISNULL(xp, 0) + @p1 WHERE user_id = @p2`
		_, err = tx.ExecContext(ctx, updateXPQ, rewardXP, userID)
		if err != nil {
			return fmt.Errorf("failed to update user xp: %w", err)
		}
	}

	if isCompleted && rewardHeart > 0 {
		updateHeartQ := `
			UPDATE users 
			SET heart_count = CASE 
				WHEN ISNULL(heart_count, 0) + @p1 > 255 THEN 255 
				ELSE ISNULL(heart_count, 0) + @p1 
			END 
			WHERE user_id = @p2`
		_, err = tx.ExecContext(ctx, updateHeartQ, rewardHeart, userID)
		if err != nil {
			return fmt.Errorf("failed to update user heart_count: %w", err)
		}
	}

	return tx.Commit()
}

func (r *repositoryImpl) GetAllUserBehaviorStats(ctx context.Context) ([]model.UserBehaviorStat, error) {
	// ปรับปรุง Query โดยเพิ่ม Subquery สำหรับนับ reflects_done ที่ JOIN ไปจนถึง Tree_Album
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id,
			
			(SELECT COUNT(*) 
			 FROM node_progress np 
			 WHERE np.user_id = u.user_id 
			   AND np.complete = 'true' 
			   AND np.updated_at >= GETDATE() - 7) as nodes_done,
			
			(SELECT COUNT(*) 
			 FROM path_enroll pe 
			 WHERE pe.user_id = u.user_id 
			   AND pe.enrollment_status = 'completed' 
			   AND pe.complete_at >= GETDATE() - 7) as paths_done,
			   
			(SELECT COUNT(r.reflect_id) 
			 FROM dbo.Reflect r
			 JOIN dbo.Tree_Node tn ON r.tree_node_id = tn.tree_node_id
			 JOIN dbo.Tree t ON tn.tree_id = t.tree_id
			 JOIN dbo.Tree_Album ta ON t.album_id = ta.album_id
			 WHERE ta.user_id = u.user_id 
			   AND r.create_at >= GETDATE() - 7) as reflects_done

		FROM users u
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repo.GetAllUserBehaviorStats failed: %w", err)
	}
	defer rows.Close()

	var stats []model.UserBehaviorStat
	for rows.Next() {
		var s model.UserBehaviorStat
		if err := rows.Scan(&s.UserID, &s.NodesDoneLast7Days, &s.PathsDoneLast7Days, &s.ReflectsDoneLast7Days); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetAllUserBehaviorStats row iteration failed: %w", err)
	}

	return stats, nil
}

func (r *repositoryImpl) DeleteExpiredUnfinishedMissions(ctx context.Context) error {
	query := `
		DELETE FROM user_mission 
		WHERE status != 'completed' AND expire_at < GETDATE()`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("repo.DeleteExpiredUnfinishedMissions failed: %w", err)
	}
	return nil
}
