package service

import (
	"context"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with hashed password
func (s *userServiceImpl) CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	if user.Email == "" {
		return "", apperror.NewBadRequest("email is required")
	}
	if user.Password == "" {
		return "", apperror.NewBadRequest("password is required")
	}
	if user.Username == "" {
		return "", apperror.NewBadRequest("username is required")
	}

	// Check if email already exists
	if existingUser, _ := s.userRepo.GetUserByEmail(ctx, user.Email); existingUser != nil {
        s.logger.WarnContext(ctx, "register failed email taken", "email", user.Email)
        return "", apperror.NewConflict("email already registered")
    }

	// Check if username already exists
	if existingUser, _ := s.userRepo.GetUserByUsername(ctx, user.Username); existingUser != nil {
        s.logger.WarnContext(ctx, "register failed username taken", "username", user.Username)
        return "", apperror.NewConflict("username already taken")
    }

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.NewInternal("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Set default values
	if user.Role == "" {
		user.Role = "user"
	}
	if user.HeartCount == 0 {
		user.HeartCount = 5 // default hearts
	}

	// Generate email verification token
	verificationToken, err := GenerateVerificationToken()
	if err != nil {
		return "", apperror.NewInternal("failed to generate verification token: %w", err)
	}
	user.IsEmailVerified = false

	// Set default profile values
	if profile.Level == 0 {
		profile.Level = 1
	}
	if profile.XP == 0 {
		profile.XP = 0
	}
	if profile.LearningStreak == 0 {
		profile.LearningStreak = 0
	}
	if profile.LearningCount == 0 {
		profile.LearningCount = 0
	}
	if profile.HourLearned == 0 {
		profile.HourLearned = 0
	}
	if profile.RankName == "" {
		profile.RankName = "Beginner"
	}

	// Create user and profile
	userID, err := s.userRepo.CreateUser(ctx, user, profile)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("user already exists")
		}
		return "", apperror.NewInternal("failed to create user: %w", err)
	}

	// Save verification token to Token table
	tokenExpiry := GetVerificationTokenExpiry()
	tokenModel := &model.Token{
		UserID:    userID,
		Token:     verificationToken,
		TokenType: model.TokenTypeEmailVerification,
		IsRevoked: false,
		ExpireAt:  tokenExpiry,
	}
	if err := s.tokenRepo.CreateToken(tokenModel); err != nil {
		// Log error but don't fail registration
		s.logger.WarnContext(ctx, "failed to save verification token", "error", err, "user_id", userID)
	}

	// Send verification email (don't fail registration if email sending fails)
	if s.emailService != nil {
		if err := s.emailService.SendVerificationEmail(user.Email, verificationToken); err != nil {
			s.logger.WarnContext(ctx, "failed to send verification email", "error", err, "email", user.Email)
		}
	}

	s.logger.InfoContext(ctx, "user registered successfully", "user_id", userID)
	return userID, nil
}

// GetUserByID retrieves user and profile by ID
func (s *userServiceImpl) GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error) {
	if id == "" {
		return nil, nil, apperror.NewBadRequest("user_id is required")
	}

	user, profile, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "get user by ID failed", "error", err, "user_id", id)
		return nil, nil, apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return nil, nil, apperror.NewNotFound("user with id '%s' not found", id)
	}

	return user, profile, nil
}

// GetUserByEmail retrieves user by email
func (s *userServiceImpl) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, apperror.NewBadRequest("email is required")
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.ErrorContext(ctx, "get user by email failed", "error", err, "email", email)
		return nil, apperror.NewInternal("failed to get user by email: %w", err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user with email '%s' not found", email)
	}

	return user, nil
}

// UpdateUser updates user information (only first_name and last_name)
func (s *userServiceImpl) UpdateUser(ctx context.Context, id string, firstName string, lastName string) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	// Check if user exists
	existingUser, _, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "update user failed", "error", err, "user_id", id)
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if existingUser == nil {
		return apperror.NewNotFound("user with id '%s' not found", id)
	}

	if err := s.userRepo.UpdateUser(ctx, id, firstName, lastName); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("username already taken")
		}
		return apperror.NewInternal("failed to update user: %w", err)
	}

	s.logger.InfoContext(ctx, "user updated successfully", "user_id", id)
	return nil
}

// DeleteUser deletes a user after password confirmation
func (s *userServiceImpl) DeleteUser(ctx context.Context, id string, password string) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}
	if password == "" {
		return apperror.NewBadRequest("password is required for account deletion")
	}

	// Get user and verify password
	user, _, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Verify password before deletion
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.WarnContext(ctx, "account deletion failed: incorrect password", "user_id", id)
		return apperror.NewUnauthorized("incorrect password")
	}

	if err := s.userRepo.DeleteUser(ctx, id); err != nil {
		return apperror.NewInternal("failed to delete user: %w", err)
	}

	s.logger.InfoContext(ctx, "user deleted successfully", "user_id", id)
	return nil
}
