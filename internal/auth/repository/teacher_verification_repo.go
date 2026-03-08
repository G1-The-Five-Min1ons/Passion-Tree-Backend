package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"passiontree/internal/auth/model"
)

func (r *repositoryImpl) GetTeacherVerificationStatus(ctx context.Context, userID string) (*model.TeacherVerificationStatus, error) {
	query := `
		IF OBJECT_ID('teacher_verification_requests', 'U') IS NULL
		BEGIN
			SELECT
				ISNULL(p.Phone_Number, '') AS phone_number,
				CASE WHEN ISNULL(p.Phone_Number, '') <> '' THEN 1 ELSE 0 END AS has_phone_number,
				'none' AS application_status
			FROM users u
			LEFT JOIN profile p ON u.user_id = p.user_id
			WHERE u.user_id = @p1
		END
		ELSE
		BEGIN
			SELECT
				ISNULL(p.Phone_Number, '') AS phone_number,
				CASE WHEN ISNULL(p.Phone_Number, '') <> '' THEN 1 ELSE 0 END AS has_phone_number,
				ISNULL(app.status, 'none') AS application_status
			FROM users u
			LEFT JOIN profile p ON u.user_id = p.user_id
			OUTER APPLY (
				SELECT TOP 1 status
				FROM teacher_verification_requests tvr
				WHERE tvr.user_id = u.user_id
				ORDER BY tvr.created_at DESC
			) app
			WHERE u.user_id = @p1
		END`

	var (
		phoneNumber       string
		hasPhoneNumberInt int
		applicationStatus string
	)

	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&phoneNumber, &hasPhoneNumberInt, &applicationStatus); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get teacher verification status failed: %w", err)
	}

	hasPhoneNumber := hasPhoneNumberInt == 1
	hasApplied := applicationStatus != "none"
	isVerified := hasPhoneNumber && applicationStatus == "approved"

	return &model.TeacherVerificationStatus{
		PhoneNumber:       phoneNumber,
		HasPhoneNumber:    hasPhoneNumber,
		HasApplied:        hasApplied,
		ApplicationStatus: applicationStatus,
		IsVerified:        isVerified,
	}, nil
}

func (r *repositoryImpl) UpsertTeacherApplication(ctx context.Context, userID, phoneNumber, reason, teachingHistory string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	updateProfileQuery := `UPDATE profile SET Phone_Number = @p1 WHERE user_id = @p2`
	if _, err := tx.ExecContext(ctx, updateProfileQuery, phoneNumber, userID); err != nil {
		return fmt.Errorf("update profile phone number failed: %w", err)
	}

	var existingRequestID string
	getRequestQuery := `
		SELECT TOP 1 CONVERT(VARCHAR(36), request_id)
		FROM teacher_verification_requests
		WHERE user_id = @p1
		ORDER BY created_at DESC`
	err = tx.QueryRowContext(ctx, getRequestQuery, userID).Scan(&existingRequestID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("check existing teacher application failed: %w", err)
	}

	if err == sql.ErrNoRows {
		insertQuery := `
			INSERT INTO teacher_verification_requests
			(request_id, user_id, phone_number, reason, teaching_history, status, created_at, updated_at)
			VALUES (@p1, @p2, @p3, @p4, @p5, 'pending', GETUTCDATE(), GETUTCDATE())`
		if _, err := tx.ExecContext(ctx, insertQuery, uuid.New().String(), userID, phoneNumber, reason, teachingHistory); err != nil {
			return fmt.Errorf("create teacher application failed: %w", err)
		}
	} else {
		updateQuery := `
			UPDATE teacher_verification_requests
			SET
				phone_number = @p1,
				reason = @p2,
				teaching_history = @p3,
				status = 'pending',
				reviewed_by = NULL,
				reviewed_at = NULL,
				updated_at = GETUTCDATE()
			WHERE request_id = @p4`
		if _, err := tx.ExecContext(ctx, updateQuery, phoneNumber, reason, teachingHistory, existingRequestID); err != nil {
			return fmt.Errorf("update teacher application failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}

func (r *repositoryImpl) ListTeacherApplications(ctx context.Context, status string) ([]model.TeacherVerificationRequest, error) {
	query := `
		SELECT
			CONVERT(VARCHAR(36), tvr.request_id) AS request_id,
			CONVERT(VARCHAR(36), tvr.user_id) AS user_id,
			u.username,
			u.email,
			u.first_name,
			u.last_name,
			tvr.phone_number,
			tvr.reason,
			tvr.teaching_history,
			tvr.status,
			ISNULL(CONVERT(VARCHAR(36), tvr.reviewed_by), '') AS reviewed_by,
			tvr.reviewed_at,
			tvr.created_at,
			tvr.updated_at
		FROM teacher_verification_requests tvr
		JOIN users u ON u.user_id = tvr.user_id
		WHERE (@p1 = '' OR tvr.status = @p1)
		ORDER BY tvr.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("list teacher applications query failed: %w", err)
	}
	defer rows.Close()

	applications := make([]model.TeacherVerificationRequest, 0)
	for rows.Next() {
		var app model.TeacherVerificationRequest
		var reviewedBy string
		if err := rows.Scan(
			&app.RequestID,
			&app.UserID,
			&app.Username,
			&app.Email,
			&app.FirstName,
			&app.LastName,
			&app.PhoneNumber,
			&app.Reason,
			&app.TeachingHistory,
			&app.Status,
			&reviewedBy,
			&app.ReviewedAt,
			&app.CreatedAt,
			&app.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("list teacher applications scan failed: %w", err)
		}
		app.ReviewedBy = reviewedBy
		applications = append(applications, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list teacher applications rows failed: %w", err)
	}

	return applications, nil
}

func (r *repositoryImpl) ReviewTeacherApplication(ctx context.Context, requestID, status, reviewedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	var userID string
	getUserIDQuery := `
		SELECT CONVERT(VARCHAR(36), user_id)
		FROM teacher_verification_requests
		WHERE request_id = @p1`
	if err := tx.QueryRowContext(ctx, getUserIDQuery, requestID).Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return fmt.Errorf("get teacher application user failed: %w", err)
	}

	updateQuery := `
		UPDATE teacher_verification_requests
		SET
			status = @p1,
			reviewed_by = @p2,
			reviewed_at = GETUTCDATE(),
			updated_at = GETUTCDATE()
		WHERE request_id = @p3`
	result, err := tx.ExecContext(ctx, updateQuery, status, reviewedBy, requestID)
	if err != nil {
		return fmt.Errorf("review teacher application failed: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	if status == "approved" {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET role = 'teacher', update_at = GETUTCDATE() WHERE user_id = @p1`, userID); err != nil {
			return fmt.Errorf("update approved teacher role failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
