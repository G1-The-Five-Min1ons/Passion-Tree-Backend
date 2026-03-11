package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/auth/model"
)

// GetAllUsers retrieves all users with their profiles (admin only)
func (r *repositoryImpl) GetAllUsers(ctx context.Context) ([]*model.UserWithProfile, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, u.username, u.email, u.first_name, u.last_name, 
			u.role, u.heart_count, u.is_email_verified, u.create_at, u.update_at,
			u.failed_attempts, u.locked_until,
			CONVERT(VARCHAR(36), p.Profile_ID) as Profile_ID, p.Avatar_URL, p.Rank_Name, 
			p.Learning_streak, p.Learning_count, p.Location, p.Bio, 
			p.Level, p.XP, p.Hour_learned
		FROM users AS u
		LEFT JOIN profile p ON u.user_id = p.user_id
		ORDER BY u.create_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all users failed: %w", err)
	}
	defer rows.Close()

	users := make([]*model.UserWithProfile, 0)
	for rows.Next() {
		var u model.User
		var p *model.Profile
		var profileID, avatarURL, rankName, location, bio sql.NullString
		var learningStreak, learningCount, level, hourLearned sql.NullInt32
		var xp sql.NullInt64
		var createdAt, updatedAt, lockedUntil sql.NullTime
		var failedAttempts int

		err := rows.Scan(
			&u.UserID, &u.Username, &u.Email, &u.FirstName, &u.LastName,
			&u.Role, &u.HeartCount, &u.IsEmailVerified, &createdAt, &updatedAt,
			&failedAttempts, &lockedUntil,
			&profileID, &avatarURL, &rankName, &learningStreak, &learningCount,
			&location, &bio, &level, &xp, &hourLearned,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user failed: %w", err)
		}

		if createdAt.Valid {
			u.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			u.UpdatedAt = updatedAt.Time
		}
		if lockedUntil.Valid {
			t := lockedUntil.Time
			u.LockedUntil = &t
		}
		u.FailedAttempts = failedAttempts

		if profileID.Valid {
			var levelPtr *int
			var xpPtr *int64
			if level.Valid {
				lv := int(level.Int32)
				levelPtr = &lv
			}
			if xp.Valid {
				xpVal := xp.Int64
				xpPtr = &xpVal
			}

			p = &model.Profile{
				ProfileID:      profileID.String,
				AvatarURL:      avatarURL.String,
				RankName:       rankName.String,
				LearningStreak: int(learningStreak.Int32),
				LearningCount:  int(learningCount.Int32),
				Location:       location.String,
				Bio:            bio.String,
				Level:          levelPtr,
				XP:             xpPtr,
				HourLearned:    int(hourLearned.Int32),
			}
		}

		users = append(users, &model.UserWithProfile{
			User:    u,
			Profile: p,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users rows failed: %w", err)
	}

	return users, nil
}

// GetDashboardStats retrieves dashboard statistics (admin only)
func (r *repositoryImpl) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	statsQuery := `
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM learning_path) AS total_paths,
			(SELECT COUNT(*) FROM path_enroll) AS total_enrollments,
			(SELECT COUNT(*) FROM reflect) AS total_reflections`

	err := r.db.QueryRowContext(ctx, statsQuery).Scan(
		&stats.TotalUsers,
		&stats.TotalPaths,
		&stats.TotalEnrollments,
		&stats.TotalReflections,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary stats: %w", err)
	}

	// Get recent activities (last 10)
	activityQuery := `
		SELECT TOP 10 activity_type, user_id, username, message, timestamp 
		FROM (
			SELECT 'user_registered' as activity_type, 
				   CONVERT(VARCHAR(36), u.user_id) as user_id,
				   u.username,
				   CONCAT(ISNULL(u.first_name, ''), ' ', ISNULL(u.last_name, ''), ' joined the platform') as message,
				   u.create_at as timestamp
			FROM users u
			UNION ALL
			SELECT 'path_created' as activity_type,
				   CONVERT(VARCHAR(36), lp.creator_id) as user_id,
				   u.username,
				   CONCAT(ISNULL(u.first_name, ''), ' ', ISNULL(u.last_name, ''), ' created learning path: ', ISNULL(lp.title, '')) as message,
				   lp.create_at as timestamp
			FROM learning_path lp
			INNER JOIN users u ON lp.creator_id = u.user_id
			UNION ALL
			SELECT 'reflection_created' as activity_type,
				   CONVERT(VARCHAR(36), ta.user_id) as user_id,
				   u.username,
				   CONCAT(ISNULL(u.first_name, ''), ' ', ISNULL(u.last_name, ''), ' published a reflection') as message,
				   r.create_at as timestamp
			FROM Reflect r
			INNER JOIN tree_node tn ON r.tree_node_id = tn.tree_node_id
			INNER JOIN tree t ON tn.tree_id = t.tree_id
			INNER JOIN tree_album ta ON t.album_id = ta.album_id
			INNER JOIN users u ON ta.user_id = u.user_id
		) activities
		ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, activityQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	defer rows.Close()

	var activities []model.Activity
	for rows.Next() {
		var a model.Activity
		err := rows.Scan(&a.Type, &a.UserID, &a.Username, &a.Message, &a.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate activity rows failed: %w", err)
	}
	stats.RecentActivities = activities

	// Get user growth for the last 12 months (including months with zero signups)
	growthQuery := `
		WITH months AS (
			SELECT
				0 AS idx,
				DATEFROMPARTS(
					YEAR(DATEADD(month, -11, GETDATE())),
					MONTH(DATEADD(month, -11, GETDATE())),
					1
				) AS month_start
			UNION ALL
			SELECT
				idx + 1,
				DATEADD(month, 1, month_start)
			FROM months
			WHERE idx < 11
		),
		user_counts AS (
			SELECT
				DATEFROMPARTS(YEAR(create_at), MONTH(create_at), 1) AS month_start,
				COUNT(*) AS user_count
			FROM users
			WHERE create_at >= DATEFROMPARTS(
				YEAR(DATEADD(month, -11, GETDATE())),
				MONTH(DATEADD(month, -11, GETDATE())),
				1
			)
			GROUP BY DATEFROMPARTS(YEAR(create_at), MONTH(create_at), 1)
		)
		SELECT
			m.idx,
			ISNULL(uc.user_count, 0) AS user_count
		FROM months m
		LEFT JOIN user_counts uc ON m.month_start = uc.month_start
		ORDER BY m.idx
		OPTION (MAXRECURSION 12)`

	rows, err = r.db.QueryContext(ctx, growthQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get user growth: %w", err)
	}
	defer rows.Close()

	growth := make([]int, 12)
	for rows.Next() {
		var idx, count int
		err := rows.Scan(&idx, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan growth: %w", err)
		}
		if idx >= 0 && idx < len(growth) {
			growth[idx] = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate growth rows failed: %w", err)
	}
	stats.UserGrowth = growth

	return stats, nil
}
