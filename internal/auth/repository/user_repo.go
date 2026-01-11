package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/auth/model"

	"github.com/google/uuid"
)

// CreateUser creates a new user with transaction support
func (r *userRepositoryImpl) CreateUser(user *model.User, profile *model.Profile) (string, error) {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	userID := uuid.New().String()

	// Insert into users table
	userQuery := `INSERT INTO users (user_id, username, email, password, first_name, last_name, role, heart_count) 
	              VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)`
	_, err = tx.ExecContext(ctx, userQuery,
		userID, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Role, user.HeartCount)
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
func (r *userRepositoryImpl) GetUserByID(id string) (*model.User, *model.Profile, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), u.user_id) as user_id, u.username, u.email, u.first_name, u.last_name, u.role, u.heart_count,
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

	err := r.db.QueryRow(query, id).Scan(
		&u.UserID, &u.Username, &u.Email, &u.FirstName, &u.LastName, &u.Role, &u.HeartCount,
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
		p.Level = int(level.Int32)
		p.XP = xp.Int64
		p.HourLearned = int(hourLearned.Int32)
		p.UserID = u.UserID
	}

	return &u, &p, nil
}

// GetUserByEmail fetches a user by email
func (r *userRepositoryImpl) GetUserByEmail(email string) (*model.User, error) {
	query := `SELECT CONVERT(VARCHAR(36), user_id) as user_id, username, email, password, first_name, last_name, role, heart_count 
	          FROM users WHERE email = @p1`
	var user model.User
	err := r.db.QueryRow(query, email).Scan(
		&user.UserID, &user.Username, &user.Email, &user.Password,
		&user.FirstName, &user.LastName, &user.Role, &user.HeartCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email failed: %w", err)
	}
	return &user, nil
}

// GetUserByUsername fetches a user by username
func (r *userRepositoryImpl) GetUserByUsername(username string) (*model.User, error) {
	query := `SELECT CONVERT(VARCHAR(36), user_id) as user_id, username, email, password, first_name, last_name, role, heart_count 
	          FROM users WHERE username = @p1`
	var user model.User
	err := r.db.QueryRow(query, username).Scan(
		&user.UserID, &user.Username, &user.Email, &user.Password,
		&user.FirstName, &user.LastName, &user.Role, &user.HeartCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username failed: %w", err)
	}
	return &user, nil
}

// UpdateUser updates user info by ID
func (r *userRepositoryImpl) UpdateUser(id string, user *model.User) error {
	query := `UPDATE users SET username=@p1, email=@p2, password=@p3, first_name=@p4, last_name=@p5, role=@p6, heart_count=@p7 
	          WHERE user_id=@p8`
	_, err := r.db.Exec(query, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Role, user.HeartCount, id)
	if err != nil {
		return fmt.Errorf("update user failed [id=%s]: %w", id, err)
	}
	return nil
}

// DeleteUser deletes a user by ID (cascade will delete profile)
func (r *userRepositoryImpl) DeleteUser(id string) error {
	_, err := r.db.Exec("DELETE FROM users WHERE user_id = @p1", id)
	if err != nil {
		return fmt.Errorf("delete user failed [id=%s]: %w", id, err)
	}
	return nil
}
