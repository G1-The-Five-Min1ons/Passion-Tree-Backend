package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"passiontree/internal/auth/model"

	"github.com/google/uuid"
)

// --- User Creation & Retrieval ---

// CreateUser creates a new user with transaction support
func (r *repositoryImpl) CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	userID := uuid.New().String()

	// Insert into users table
	userQuery := `INSERT INTO users (user_id, username, email, password, first_name, last_name, role, heart_count, is_email_verified, create_at, update_at) 
	              VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, GETDATE(), GETDATE())`
	_, err = tx.ExecContext(ctx, userQuery,
		userID, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Role, user.HeartCount,
		user.IsEmailVerified)
	if err != nil {
		return "", fmt.Errorf("insert users failed: %w", err)
	}

	// Insert into profile table
	profileID := uuid.New().String()
	profileQuery := `INSERT INTO profile (Profile_ID, Avatar_URL, Rank_Name, Learning_streak, Learning_count, Location, Bio, Level, XP, Hour_learned, user_id) 
	                 VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11)`
	_, err = tx.ExecContext(ctx, profileQuery,
		profileID, profile.AvatarURL, profile.RankName, profile.LearningStreak, profile.LearningCount,
		profile.Location, profile.Bio, profile.Level, profile.XP, profile.HourLearned, userID)
	if err != nil {
		return "", fmt.Errorf("insert profile failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit transaction failed: %w", err)
	}

	return userID, nil
}

// GetUserByID fetches a user and profile by ID
func (r *repositoryImpl) GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, u.username, u.email, u.password, u.first_name, u.last_name, u.role, u.heart_count,
			u.is_email_verified, u.require_2fa_next_login, u.create_at, u.update_at,
			CONVERT(VARCHAR(36), p.Profile_ID) as Profile_ID, p.Avatar_URL, p.Rank_Name, p.Learning_streak, p.Learning_count, 
			p.Location, p.Bio, p.Phone_Number, p.Time_Zone, p.Date_Format, p.Level, p.XP, p.Hour_learned
		FROM users AS u
		LEFT JOIN profile p ON u.user_id = p.user_id
		WHERE u.user_id = @p1`

	var u model.User
	var p *model.Profile
	var profileID, avatarURL, rankName, location, bio, phoneNumber, timeZone, dateFormat sql.NullString
	var learningStreak, learningCount, level, hourLearned sql.NullInt32
	var xp sql.NullInt64
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.UserID, &u.Username, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.HeartCount,
		&u.IsEmailVerified, &u.Require2FANextLogin, &createdAt, &updatedAt,
		&profileID, &avatarURL, &rankName, &learningStreak, &learningCount,
		&location, &bio, &phoneNumber, &timeZone, &dateFormat, &level, &xp, &hourLearned,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("get user by id failed: %w", err)
	}

	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}

	if profileID.Valid {
		p = &model.Profile{}
		p.ProfileID = profileID.String
		p.AvatarURL = avatarURL.String
		p.RankName = rankName.String
		p.LearningStreak = int(learningStreak.Int32)
		p.LearningCount = int(learningCount.Int32)
		p.Location = location.String
		p.Bio = bio.String
		p.PhoneNumber = phoneNumber.String
		p.TimeZone = timeZone.String
		p.DateFormat = dateFormat.String
		if level.Valid {
			levelValue := int(level.Int32)
			p.Level = &levelValue
		}
		if xp.Valid {
			xpValue := xp.Int64
			p.XP = &xpValue
		}
		p.HourLearned = int(hourLearned.Int32)
		p.UserID = u.UserID
	}

	return &u, p, nil
}

// --- Auth Helpers ---

func (r *repositoryImpl) fetchUser(ctx context.Context, query string, value interface{}) (*model.User, error) {
	var user model.User
	var lockedUntil sql.NullTime

	err := r.db.QueryRowContext(ctx, query, value).Scan(
		&user.UserID, &user.Username, &user.Email, &user.Password,
		&user.FirstName, &user.LastName, &user.Role, &user.HeartCount,
		&user.IsEmailVerified, &user.Require2FANextLogin, &user.FailedAttempts, &lockedUntil,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}

	return &user, nil
}

func (r *repositoryImpl) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT 
				CONVERT(VARCHAR(36), user_id) as user_id, username, email, password, 
				first_name, last_name, role, heart_count, is_email_verified, require_2fa_next_login,
				failed_attempts, locked_until 
			  FROM users WHERE email = @p1`
	return r.fetchUser(ctx, query, email)
}

func (r *repositoryImpl) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT 
				CONVERT(VARCHAR(36), user_id) as user_id, username, email, password, 
				first_name, last_name, role, heart_count, is_email_verified, require_2fa_next_login,
				failed_attempts, locked_until 
			  FROM users WHERE username = @p1`
	return r.fetchUser(ctx, query, username)
}

// --- Updates & Deletion ---

func (r *repositoryImpl) UpdateUser(ctx context.Context, id string, username string, firstName string, lastName string, role string) error {
	var updates []string
	var args []interface{}
	paramID := 1

	if username != "" {
		updates = append(updates, fmt.Sprintf("username=@p%d", paramID))
		args = append(args, username)
		paramID++
	}
	if firstName != "" {
		updates = append(updates, fmt.Sprintf("first_name=@p%d", paramID))
		args = append(args, firstName)
		paramID++
	}
	if lastName != "" {
		updates = append(updates, fmt.Sprintf("last_name=@p%d", paramID))
		args = append(args, lastName)
		paramID++
	}
	if role != "" {
		updates = append(updates, fmt.Sprintf("role=@p%d", paramID))
		args = append(args, role)
		paramID++
	}

	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE users SET " + strings.Join(updates, ", ") + fmt.Sprintf(", update_at=GETDATE() WHERE user_id=@p%d", paramID)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repositoryImpl) UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error {
	if userID == "" || profile == nil {
		return fmt.Errorf("userID and profile are required")
	}

	var updates []string
	var args []interface{}
	paramID := 1

	// Dynamic query building
	if profile.AvatarURL != "" {
		updates = append(updates, fmt.Sprintf("Avatar_URL=@p%d", paramID))
		args = append(args, profile.AvatarURL)
		paramID++
	}
	if profile.RankName != "" {
		updates = append(updates, fmt.Sprintf("Rank_Name=@p%d", paramID))
		args = append(args, profile.RankName)
		paramID++
	}
	if profile.Location != "" {
		updates = append(updates, fmt.Sprintf("Location=@p%d", paramID))
		args = append(args, profile.Location)
		paramID++
	}
	if profile.Bio != "" {
		updates = append(updates, fmt.Sprintf("Bio=@p%d", paramID))
		args = append(args, profile.Bio)
		paramID++
	}
	if profile.PhoneNumber != "" {
		updates = append(updates, fmt.Sprintf("Phone_Number=@p%d", paramID))
		args = append(args, profile.PhoneNumber)
		paramID++
	}
	if profile.TimeZone != "" {
		updates = append(updates, fmt.Sprintf("Time_Zone=@p%d", paramID))
		args = append(args, profile.TimeZone)
		paramID++
	}
	if profile.DateFormat != "" {
		updates = append(updates, fmt.Sprintf("Date_Format=@p%d", paramID))
		args = append(args, profile.DateFormat)
		paramID++
	}
	if profile.Level != nil {
		updates = append(updates, fmt.Sprintf("Level=@p%d", paramID))
		args = append(args, *profile.Level)
		paramID++
	}
	if profile.XP != nil {
		updates = append(updates, fmt.Sprintf("XP=@p%d", paramID))
		args = append(args, *profile.XP)
		paramID++
	}
	// ... add other fields as needed ...

	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE profile SET " + strings.Join(updates, ", ") + fmt.Sprintf(" WHERE user_id=@p%d", paramID)
	args = append(args, userID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Check if user exists
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id '%s' not found", userID)
	}

	return nil
}

func (r *repositoryImpl) DeleteUser(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.ExecContext(ctx, "DELETE FROM profile WHERE user_id = @p1", id)
	_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE user_id = @p1", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) SetAccountDeactivatedUntil(ctx context.Context, userID string, until time.Time) error {
    query := `
        MERGE INTO settings AS target
        USING (SELECT @p1 AS user_id, @p2 AS [key]) AS source
        ON (target.user_id = source.user_id AND target.[key] = source.[key])
        WHEN MATCHED THEN
            UPDATE SET [value] = @p3, updated_at = GETDATE()
        WHEN NOT MATCHED THEN
            INSERT (id, user_id, [key], [value], created_at, updated_at)
            VALUES (NEWID(), @p1, @p2, @p3, GETDATE(), GETDATE());
    `

    value := until.UTC().Format(time.RFC3339)
    _, err := r.db.ExecContext(ctx, query, userID, "account_deactivated_until", value)
    return err
}

func (r *repositoryImpl) GetAccountDeactivatedUntil(ctx context.Context, userID string) (*time.Time, error) {
	query := `SELECT [value] FROM settings WHERE user_id = @p1 AND [key] = @p2`

	var raw string
	err := r.db.QueryRowContext(ctx, query, userID, "account_deactivated_until").Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func (r *repositoryImpl) ClearAccountDeactivatedUntil(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM settings WHERE user_id = @p1 AND [key] = @p2", userID, "account_deactivated_until")
	return err
}

// --- Password & Security ---

// UpdatePassword includes rowsAffected check for security
func (r *repositoryImpl) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	query := `UPDATE users SET password = @p1 WHERE user_id = @p2`
	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or no changes made")
	}
	return nil
}

// ResetPasswordWithToken implements Global Logout (Revoke all refresh tokens)
func (r *repositoryImpl) ResetPasswordWithToken(ctx context.Context, userID string, hashedPassword string, tokenID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update Password
	if _, err := tx.ExecContext(ctx, `UPDATE users SET password = @p1 WHERE user_id = @p2`, hashedPassword, userID); err != nil {
		return err
	}
	// 2. Revoke Reset Token
	if _, err := tx.ExecContext(ctx, `UPDATE Token SET is_revoke = 1 WHERE token_id = @p1`, tokenID); err != nil {
		return err
	}
	// 3. Security: Global Logout
	revokeQuery := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	if _, err := tx.ExecContext(ctx, revokeQuery, userID, model.TokenTypeRefresh); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) ChangePasswordAndRevokeSessions(ctx context.Context, userID string, hashedPassword string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `UPDATE users SET password = @p1 WHERE user_id = @p2`, hashedPassword, userID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	revokeQuery := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	_, err = tx.ExecContext(ctx, revokeQuery, userID, model.TokenTypeRefresh)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) SetRequire2FANextLogin(ctx context.Context, userID string, require2FA bool) error {
	query := `UPDATE users SET require_2fa_next_login = @p1, update_at = GETDATE() WHERE user_id = @p2`
	_, err := r.db.ExecContext(ctx, query, require2FA, userID)
	return err
}

// --- Verification & Failed Logins ---

func (r *repositoryImpl) VerifyEmailWithToken(ctx context.Context, userID string, tokenValue string, tokenType string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE users SET is_email_verified = 1, update_at = GETDATE() WHERE user_id = @p1`, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE Token SET is_revoke = 1 WHERE token = @p1 AND token_type = @p2`, tokenValue, tokenType); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) ReplaceVerificationToken(ctx context.Context, userID string, newToken *model.Token) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.ExecContext(ctx, `DELETE FROM Token WHERE user_id = @p1 AND token_type = @p2`, userID, model.TokenTypeEmailVerification)

	insertQuery := `INSERT INTO Token (user_id, token, token_type, is_revoke, expire_at) VALUES (@p1, @p2, @p3, @p4, @p5)`
	_, err = tx.ExecContext(ctx, insertQuery, newToken.UserID, newToken.Token, newToken.TokenType, newToken.IsRevoked, newToken.ExpireAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) UpdateFailedLogin(ctx context.Context, userID string, lockDuration time.Duration) (int, error) {
	query := `
		UPDATE users 
		SET failed_attempts = failed_attempts + 1,
			locked_until = CASE 
				WHEN failed_attempts + 1 >= 5 THEN @p1 
				ELSE locked_until 
			END
		OUTPUT INSERTED.failed_attempts
		WHERE user_id = @p2`

	lockTime := time.Now().UTC().Add(lockDuration)
	var newAttempts int
	err := r.db.QueryRowContext(ctx, query, lockTime, userID).Scan(&newAttempts)
	return newAttempts, err
}

func (r *repositoryImpl) ResetFailedLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE user_id = @p1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *repositoryImpl) UpdateEmailVerified(ctx context.Context, userID string, isVerified bool) error {
	query := `
		UPDATE users 
		SET is_email_verified = @p1, 
		    update_at = GETDATE() 
		WHERE user_id = @p2`

	result, err := r.db.ExecContext(ctx, query, isVerified, userID)
	if err != nil {
		return fmt.Errorf("failed to update email verification status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}

	return nil
}

// DeactivateAccountWithTokenRevoke deactivates account and revokes all refresh tokens atomically
func (r *repositoryImpl) DeactivateAccountWithTokenRevoke(ctx context.Context, userID string, until time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Set account deactivated until
	deactivateQuery := `
		IF EXISTS (SELECT 1 FROM settings WHERE user_id = @p1 AND [key] = @p2)
		BEGIN
			UPDATE settings
			SET [value] = @p3, updated_at = GETDATE()
			WHERE user_id = @p1 AND [key] = @p2
		END
		ELSE
		BEGIN
			INSERT INTO settings (id, user_id, [key], [value], created_at, updated_at)
			VALUES (NEWID(), @p1, @p2, @p3, GETDATE(), GETDATE())
		END
	`

	value := until.UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx, deactivateQuery, userID, "account_deactivated_until", value); err != nil {
		return fmt.Errorf("set account deactivated until failed: %w", err)
	}

	// Revoke all refresh tokens
	revokeQuery := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	if _, err := tx.ExecContext(ctx, revokeQuery, userID, model.TokenTypeRefresh); err != nil {
		return fmt.Errorf("revoke all user tokens failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
