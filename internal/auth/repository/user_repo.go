package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"passiontree/internal/auth/model"

	"github.com/google/uuid"
)

// CreateUser creates a new user with transaction support
func (r *userRepositoryImpl) CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	userID := uuid.New().String()

	// Insert into users table
	userQuery := `INSERT INTO users (user_id, username, email, password, first_name, last_name, role, heart_count, is_email_verified) 
	              VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9)`
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
func (r *userRepositoryImpl) GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, u.username, u.email, u.password, u.first_name, u.last_name, u.role, u.heart_count,
			u.is_email_verified,
			CONVERT(VARCHAR(36), p.Profile_ID) as Profile_ID, p.Avatar_URL, p.Rank_Name, p.Learning_streak, p.Learning_count, 
			p.Location, p.Bio, p.Level, p.XP, p.Hour_learned
		FROM users AS u
		LEFT JOIN profile p ON u.user_id = p.user_id
		WHERE u.user_id = @p1`

	var u model.User
	var p model.Profile
	var profileID, avatarURL, rankName, location, bio sql.NullString
	var learningStreak, learningCount, level, hourLearned sql.NullInt32
	var xp sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.UserID, &u.Username, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.HeartCount,
		&u.IsEmailVerified,
		&profileID, &avatarURL, &rankName, &learningStreak, &learningCount,
		&location, &bio, &level, &xp, &hourLearned,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("get user by id failed: %w", err)
	}

	// Map nullable fields to profile
	if profileID.Valid {
		p.ProfileID = profileID.String
		p.AvatarURL = avatarURL.String
		p.RankName = rankName.String
		p.LearningStreak = int(learningStreak.Int32)
		p.LearningCount = int(learningCount.Int32)
		p.Location = location.String
		p.Bio = bio.String
		
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

	return &u, &p, nil
}

func (r *userRepositoryImpl) UpdateFailedLogin(ctx context.Context, userID string, lockDuration time.Duration) (int, error) {
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
    if err != nil {
        return 0, fmt.Errorf("atomic update failed_login failed: %w", err)
    }

    return newAttempts, nil
}

func (r *userRepositoryImpl) ResetFailedLogin(ctx context.Context, userID string) error {
    query := `UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE user_id = @p1`
    _, err := r.db.ExecContext(ctx, query, userID)
    if err != nil {
        return fmt.Errorf("reset failed login failed [user_id=%s]: %w", userID, err)
    }
    return nil
}

// fetchUser is a private helper to reduce duplication of GetUserByEmail and GetUserByUsername
func (r *userRepositoryImpl) fetchUser(ctx context.Context, query string, value interface{}) (*model.User, error) {
    var user model.User
    var lockedUntil sql.NullTime

    err := r.db.QueryRowContext(ctx, query, value).Scan(
        &user.UserID, &user.Username, &user.Email, &user.Password,
        &user.FirstName, &user.LastName, &user.Role, &user.HeartCount,
        &user.IsEmailVerified, &user.FailedAttempts, &lockedUntil,
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

// GetUserByEmail fetches a user by email
func (r *userRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
    query := `SELECT 
                CONVERT(VARCHAR(36), user_id) as user_id, username, email, password, 
                first_name, last_name, role, heart_count, is_email_verified,
                failed_attempts, locked_until 
              FROM users WHERE email = @p1`
    
    user, err := r.fetchUser(ctx, query, email)
    if err != nil {
        return nil, fmt.Errorf("get user by email failed: %w", err)
    }
    return user, nil
}

// GetUserByUsername fetches a user by username
func (r *userRepositoryImpl) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
    query := `SELECT 
                CONVERT(VARCHAR(36), user_id) as user_id, username, email, password, 
                first_name, last_name, role, heart_count, is_email_verified,
                failed_attempts, locked_until 
              FROM users WHERE username = @p1`

    user, err := r.fetchUser(ctx, query, username)
    if err != nil {
        return nil, fmt.Errorf("get user by username failed: %w", err)
    }
    return user, nil
}

// UpdateUser updates user info by ID (only first_name and last_name)
func (r *userRepositoryImpl) UpdateUser(ctx context.Context, id string, firstName string, lastName string) error {
	query := `UPDATE users SET first_name=@p1, last_name=@p2 WHERE user_id=@p3`
	_, err := r.db.ExecContext(ctx, query, firstName, lastName, id)
	if err != nil {
		return fmt.Errorf("update user failed [id=%s]: %w", id, err)
	}
	return nil
}

// UpdateProfile updates profile info by user ID (Partial Update support)
func (r *userRepositoryImpl) UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error {
    if userID == "" || profile == nil {
        return fmt.Errorf("userID and profile are required for update")
    }

    // Create Dynamic Query
    var updates []string
    var args []interface{}
    paramID := 1 // เริ่มต้นที่ @p1

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
    if profile.LearningStreak > 0 {
        updates = append(updates, fmt.Sprintf("Learning_streak=@p%d", paramID))
        args = append(args, profile.LearningStreak)
        paramID++
    }
    if profile.LearningCount > 0 {
        updates = append(updates, fmt.Sprintf("Learning_count=@p%d", paramID))
        args = append(args, profile.LearningCount)
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
    if profile.HourLearned > 0 {
        updates = append(updates, fmt.Sprintf("Hour_learned=@p%d", paramID))
        args = append(args, profile.HourLearned)
        paramID++
    }


    if len(updates) == 0 {
        slog.Warn("UpdateProfile called with no fields to update", "userID", userID)
        return nil
    }

    // Combine Query หลัก
    baseQuery := "UPDATE profile SET "
    finalQuery := baseQuery + strings.Join(updates, ", ") + fmt.Sprintf(" WHERE user_id=@p%d", paramID)
    args = append(args, userID)

    _, err := r.db.ExecContext(ctx, finalQuery, args...)
    if err != nil {
        return fmt.Errorf("partial update profile failed [user_id=%s]: %w", userID, err)
    }

    return nil
}

// DeleteUser deletes a user by ID (must delete profile first due to FK constraint)
func (r *userRepositoryImpl) DeleteUser(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Delete profile first to avoid FK constraint violation
	_, err = tx.ExecContext(ctx, "DELETE FROM profile WHERE user_id = @p1", id)
	if err != nil {
        return fmt.Errorf("delete profile failed: %w", err)
    }

	// Then delete user
	_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE user_id = @p1", id)
	if err != nil {
		return fmt.Errorf("delete user failed [id=%s]: %w", id, err)
	}

	return tx.Commit()
}

// UpdateEmailVerified updates the email verification status for a user
func (r *userRepositoryImpl) UpdateEmailVerified(ctx context.Context, userID string, isVerified bool) error {
	query := `UPDATE users SET is_email_verified=@p1 WHERE user_id=@p2`
	_, err := r.db.ExecContext(ctx, query, isVerified, userID)
	if err != nil {
		return fmt.Errorf("update email verified failed [user_id=%s]: %w", userID, err)
	}
	return nil
}

// VerifyEmailAndRevokeToken verifies email and revokes verification token in a transaction
func (r *userRepositoryImpl) VerifyEmailAndRevokeToken(ctx context.Context, userID string, tokenID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update email verification status
	updateEmailQuery := `UPDATE users SET is_email_verified=@p1 WHERE user_id=@p2`
	if _, err := tx.ExecContext(ctx, updateEmailQuery, true, userID); err != nil {
		return fmt.Errorf("failed to update email verified: %w", err)
	}

	// Revoke the verification token
	revokeTokenQuery := `UPDATE Token SET is_revoke = 1 WHERE token_id = @p1`
	if _, err := tx.ExecContext(ctx, revokeTokenQuery, tokenID); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdatePassword updates user password
func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	query := `UPDATE users SET password = @p1 WHERE user_id = @p2`
	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("update password failed [user_id=%s]: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ResetPasswordWithToken resets password and revokes token in a transaction
func (r *userRepositoryImpl) ResetPasswordWithToken(ctx context.Context, userID string, hashedPassword string, tokenID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update password
	updatePasswordQuery := `UPDATE users SET password = @p1 WHERE user_id = @p2`
	if _, err := tx.ExecContext(ctx, updatePasswordQuery, hashedPassword, userID); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke the reset token
	revokeTokenQuery := `UPDATE Token SET is_revoke = 1 WHERE token_id = @p1`
	if _, err := tx.ExecContext(ctx, revokeTokenQuery, tokenID); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	// Revoke all refresh tokens (invalidate all active sessions)
	revokeRefreshTokensQuery := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	if _, err := tx.ExecContext(ctx, revokeRefreshTokensQuery, userID, model.TokenTypeRefresh); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ChangePasswordAndRevokeSessions updates password and revokes all refresh tokens in a transaction
func (r *userRepositoryImpl) ChangePasswordAndRevokeSessions(ctx context.Context, userID string, hashedPassword string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update password
	updatePasswordQuery := `UPDATE users SET password = @p1 WHERE user_id = @p2`
	result, err := tx.ExecContext(ctx, updatePasswordQuery, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Revoke all refresh tokens (invalidate all active sessions)
	revokeRefreshTokensQuery := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	if _, err := tx.ExecContext(ctx, revokeRefreshTokensQuery, userID, model.TokenTypeRefresh); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
