package repository

import (
	"context"
	"database/sql"

	"passiontree/internal/connection"
	workerModel "passiontree/internal/worker/model"
)

type SQLEmailNotificationDataProvider struct {
	db connection.Database
}

func NewSQLEmailNotificationDataProvider(db connection.Database) *SQLEmailNotificationDataProvider {
	return &SQLEmailNotificationDataProvider{db: db}
}

func (p *SQLEmailNotificationDataProvider) DailyReminderRecipients(ctx context.Context, settingKey string) ([]workerModel.NotificationRecipient, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, 
			u.email, 
			ISNULL(u.first_name, '') as first_name
		FROM users u
		JOIN settings s ON s.user_id = u.user_id
		WHERE s.[key] = @p1 
		  AND LOWER(s.[value]) = 'true' 
		  AND u.is_email_verified = 1
		  AND NOT EXISTS (
			  SELECT 1 FROM node_progress np
			  WHERE np.user_id = u.user_id
				AND np.updated_at >= DATEADD(day, -1, GETDATE())
		  )`

	rows, err := p.db.GetDB().QueryContext(ctx, query, settingKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows, func(rows *sql.Rows) (workerModel.NotificationRecipient, error) {
		var row workerModel.NotificationRecipient
		err := rows.Scan(&row.UserID, &row.Email, &row.FirstName)
		return row, err
	})
}

func (p *SQLEmailNotificationDataProvider) WeeklyProgressRows(ctx context.Context, settingKey string) ([]workerModel.WeeklyProgressRow, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, 
			u.email, 
			ISNULL(u.first_name, '') as first_name,
			(
				SELECT COUNT(*) 
				FROM node_progress np 
				WHERE np.user_id = u.user_id 
				  AND np.complete = 'true' 
				  AND np.updated_at >= DATEADD(day, -7, GETDATE())
			) as completed_nodes,
			(
				SELECT COUNT(DISTINCT CONVERT(date, np.updated_at)) 
				FROM node_progress np 
				WHERE np.user_id = u.user_id 
				  AND np.updated_at >= DATEADD(day, -7, GETDATE())
			) as active_days
		FROM users u
		JOIN settings s ON s.user_id = u.user_id
		WHERE s.[key] = @p1 
		  AND LOWER(s.[value]) = 'true' 
		  AND u.is_email_verified = 1`

	rows, err := p.db.GetDB().QueryContext(ctx, query, settingKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows, func(rows *sql.Rows) (workerModel.WeeklyProgressRow, error) {
		var row workerModel.WeeklyProgressRow
		err := rows.Scan(&row.UserID, &row.Email, &row.FirstName, &row.CompletedNodes, &row.ActiveDays)
		return row, err
	})
}

func (p *SQLEmailNotificationDataProvider) RecommendationNewPathCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM learning_path WHERE create_at >= DATEADD(day, -7, GETDATE())`
	
	var newPathCount int
	err := p.db.GetDB().QueryRowContext(ctx, query).Scan(&newPathCount)
	if err != nil {
		return 0, err
	}

	return newPathCount, nil
}

func (p *SQLEmailNotificationDataProvider) RecipientsBySetting(ctx context.Context, settingKey string) ([]workerModel.NotificationRecipient, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, 
			u.email, 
			ISNULL(u.first_name, '') as first_name
		FROM users u
		JOIN settings s ON s.user_id = u.user_id
		WHERE s.[key] = @p1 
		  AND LOWER(s.[value]) = 'true' 
		  AND u.is_email_verified = 1`

	rows, err := p.db.GetDB().QueryContext(ctx, query, settingKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows, func(rows *sql.Rows) (workerModel.NotificationRecipient, error) {
		var row workerModel.NotificationRecipient
		err := rows.Scan(&row.UserID, &row.Email, &row.FirstName)
		return row, err
	})
}

func (p *SQLEmailNotificationDataProvider) CommentNotificationRows(ctx context.Context, settingKey string) ([]workerModel.CommentNotificationRow, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, 
			u.email, 
			ISNULL(u.first_name, '') as first_name, 
			COUNT(c.comment_id) as comment_count
		FROM users u
		JOIN settings s ON s.user_id = u.user_id
		JOIN learning_path lp ON lp.creator_ID = u.user_id
		JOIN Node_Comment c ON c.path_id = lp.path_id
		WHERE s.[key] = @p1 
		  AND LOWER(s.[value]) = 'true' 
		  AND u.is_email_verified = 1
		  AND c.create_at >= DATEADD(day, -1, GETDATE())
		  AND c.user_id <> u.user_id
		GROUP BY u.user_id, u.email, u.first_name`

	rows, err := p.db.GetDB().QueryContext(ctx, query, settingKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows, func(rows *sql.Rows) (workerModel.CommentNotificationRow, error) {
		var row workerModel.CommentNotificationRow
		err := rows.Scan(&row.UserID, &row.Email, &row.FirstName, &row.NewComments)
		return row, err
	})
}

// scanRows เป็น helper สำหรับลด code ซ้ำซ้อนในการวน loop rows
func scanRows[T any](rows *sql.Rows, scanFn func(rows *sql.Rows) (T, error)) ([]T, error) {
	result := make([]T, 0)
	for rows.Next() {
		item, err := scanFn(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}