package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/auth/model"

	"github.com/google/uuid"
)

// GetUserByProvider retrieves a user by their social auth provider and provider user ID
func (r *socialAuthRepositoryImpl) GetUserByProvider(ctx context.Context, provider, providerUserID string) (*model.User, error) {
	query := `
		SELECT
			user_id, username, email, first_name, last_name, role,
			heart_count, is_email_verified, created_at, updated_at,
			auth_provider, provider_user_id
		FROM users
		WHERE auth_provider = @p1 AND provider_user_id = @p2
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, provider, providerUserID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.HeartCount,
		&user.IsEmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.AuthProvider,
		&user.ProviderUserID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by provider: %w", err)
	}

	return &user, nil
}

// CreateSocialUser creates a new user from social auth provider
func (r *socialAuthRepositoryImpl) CreateSocialUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate new user ID
	userID := uuid.New().String()
	user.UserID = userID

	// Set default role if not provided
	if user.Role == "" {
		user.Role = model.RoleStudent
	}

	// Social auth users are automatically email verified
	user.IsEmailVerified = true

	// Insert user
	userQuery := `
		INSERT INTO users (
			user_id, username, email, password, first_name, last_name, 
			role, heart_count, is_email_verified, auth_provider, provider_user_id
		) VALUES (
			@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11
		)
	`

	_, err = tx.ExecContext(ctx, userQuery,
		userID,
		user.Username,
		user.Email,
		"", // No password for social auth
		user.FirstName,
		user.LastName,
		user.Role,
		user.HeartCount,
		user.IsEmailVerified,
		user.AuthProvider,
		user.ProviderUserID,
	)

	if err != nil {
		return "", fmt.Errorf("failed to create social user: %w", err)
	}

	// Insert profile if provided
	if profile != nil {
		profile.UserID = userID
		profileQuery := `
			INSERT INTO profiles (
				user_id, bio, location, avatar_url
			) VALUES (
				@p1, @p2, @p3, @p4
			)
		`

		_, err = tx.ExecContext(ctx, profileQuery,
			profile.UserID,
			profile.Bio,
			profile.Location,
			profile.AvatarURL,
		)

		if err != nil {
			return "", fmt.Errorf("failed to create profile: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userID, nil
}

// LinkSocialAccount links a social account to an existing user
func (r *socialAuthRepositoryImpl) LinkSocialAccount(ctx context.Context, userID, provider, providerUserID string) error {
	query := `
		UPDATE users
		SET auth_provider = @p1, provider_user_id = @p2, updated_at = GETDATE()
		WHERE user_id = @p3
	`

	result, err := r.db.ExecContext(ctx, query, provider, providerUserID, userID)
	if err != nil {
		return fmt.Errorf("failed to link social account: %w", err)
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

// UpdateSocialUserInfo updates user information from OAuth provider
func (r *socialAuthRepositoryImpl) UpdateSocialUserInfo(ctx context.Context, userID string, userInfo *model.OAuthUserInfo) error {
	query := `
		UPDATE users
		SET 
			first_name = @p1,
			last_name = @p2,
			email = @p3,
			is_email_verified = 1,
			updated_at = GETDATE()
		WHERE user_id = @p4
	`

	result, err := r.db.ExecContext(ctx, query,
		userInfo.FirstName,
		userInfo.LastName,
		userInfo.Email,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update social user info: %w", err)
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

// UpsertSocialUserProfile creates or updates user profile
func (r *socialAuthRepositoryImpl) UpsertSocialUserProfile(ctx context.Context, userID string, profile *model.Profile) error {
	if profile == nil {
		return nil
	}

	// Check if profile exists
	checkQuery := `SELECT COUNT(*) FROM profiles WHERE user_id = @p1`
	var count int
	err := r.db.QueryRowContext(ctx, checkQuery, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check profile existence: %w", err)
	}

	if count > 0 {
		// Update existing profile (only update avatar_url from social auth)
		updateQuery := `
			UPDATE profiles
			SET 
				avatar_url = CASE 
					WHEN @p1 IS NOT NULL AND @p1 != '' THEN @p1 
					ELSE avatar_url 
				END,
				updated_at = GETDATE()
			WHERE user_id = @p2
		`
		_, err = r.db.ExecContext(ctx, updateQuery, profile.AvatarURL, userID)
		if err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
	} else {
		// Create new profile
		insertQuery := `
			INSERT INTO profiles (user_id, bio, location, avatar_url)
			VALUES (@p1, @p2, @p3, @p4)
		`
		_, err = r.db.ExecContext(ctx, insertQuery,
			userID,
			profile.Bio,
			profile.Location,
			profile.AvatarURL,
		)
		if err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}
	}

	return nil
}
