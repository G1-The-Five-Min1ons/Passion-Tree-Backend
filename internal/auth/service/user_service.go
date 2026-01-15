package service

import (
	"database/sql"
	"fmt"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with hashed password
func (s *userServiceImpl) CreateUser(user *model.User, profile *model.Profile) (string, error) {
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
	existingUser, err := s.userRepo.GetUserByEmail(user.Email)
	if err != nil && err != sql.ErrNoRows {
		return "", apperror.NewInternal(err)
	}
	if existingUser != nil {
		return "", apperror.NewConflict("email already registered")
	}

	// Check if username already exists
	existingUsername, err := s.userRepo.GetUserByUsername(user.Username)
	if err != nil && err != sql.ErrNoRows {
		return "", apperror.NewInternal(err)
	}
	if existingUsername != nil {
		return "", apperror.NewConflict("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.NewInternal(fmt.Errorf("failed to hash password: %w", err))
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
		return "", apperror.NewInternal(fmt.Errorf("failed to generate verification token: %w", err))
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
	userID, err := s.userRepo.CreateUser(user, profile)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("user with this email or username already exists")
		}
		return "", apperror.NewInternal(err)
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
		fmt.Printf("Warning: failed to save verification token: %v\n", err)
	}

	// Send verification email (don't fail registration if email sending fails)
	if s.emailService != nil {
		if err := s.emailService.SendVerificationEmail(user.Email, verificationToken); err != nil {
			// Log error but don't fail registration
			fmt.Printf("Warning: failed to send verification email to %s: %v\n", user.Email, err)
		}
	}

	return userID, nil
}

// GetUserByID retrieves user and profile by ID
func (s *userServiceImpl) GetUserByID(id string) (*model.User, *model.Profile, error) {
	if id == "" {
		return nil, nil, apperror.NewBadRequest("user_id is required")
	}

	user, profile, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, nil, apperror.NewInternal(err)
	}
	if user == nil {
		return nil, nil, apperror.NewNotFound("user with id '%s' not found", id)
	}

	return user, profile, nil
}

// GetUserByEmail retrieves user by email
func (s *userServiceImpl) GetUserByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, apperror.NewBadRequest("email is required")
	}

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user with email '%s' not found", email)
	}

	return user, nil
}

// UpdateUser updates user information (only first_name and last_name)
func (s *userServiceImpl) UpdateUser(id string, firstName string, lastName string) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	// Check if user exists
	existingUser, _, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if existingUser == nil {
		return apperror.NewNotFound("user with id '%s' not found", id)
	}

	if err := s.userRepo.UpdateUser(id, firstName, lastName); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("email or username already exists")
		}
		return apperror.NewInternal(err)
	}

	return nil
}

// DeleteUser deletes a user after password confirmation
func (s *userServiceImpl) DeleteUser(id string, password string) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}
	if password == "" {
		return apperror.NewBadRequest("password is required for account deletion")
	}

	// Get user and verify password
	user, _, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Verify password before deletion
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return apperror.NewUnauthorized("invalid password")
	}

	if err := s.userRepo.DeleteUser(id); err != nil {
		return apperror.NewInternal(err)
	}

	return nil
}

// Login authenticates user and returns token
// identifier can be either username or email
func (s *userServiceImpl) Login(identifier string, password string) (string, error) {
	if identifier == "" {
		return "", apperror.NewBadRequest("username or email is required")
	}
	if password == "" {
		return "", apperror.NewBadRequest("password is required")
	}

	// Try to find user by email first, then by username
	var user *model.User
	var err error

	// Check if identifier is email (contains @)
	if contains := false; len(identifier) > 0 {
		for _, ch := range identifier {
			if ch == '@' {
				contains = true
				break
			}
		}
		if contains {
			user, err = s.userRepo.GetUserByEmail(identifier)
		} else {
			user, err = s.userRepo.GetUserByUsername(identifier)
		}
	}

	if err != nil {
		return "", apperror.NewInternal(err)
	}
	if user == nil {
		return "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// Generate JWT token
	jwtService := jwt.NewService()
	token, err := jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", apperror.NewInternal(fmt.Errorf("failed to generate token: %w", err))
	}

	return token, nil
}

// ValidateToken validates JWT token and returns user
func (s *userServiceImpl) ValidateToken(token string) (*model.User, error) {
	if token == "" {
		return nil, apperror.NewUnauthorized("token is required")
	}

	// Validate JWT token
	jwtService := jwt.NewService()
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		return nil, apperror.NewUnauthorized("invalid or expired token")
	}

	// Get user from database
	user, _, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user not found")
	}

	return user, nil
}

// VerifyEmail verifies a user's email using verification token
func (s *userServiceImpl) VerifyEmail(token string) error {
	if token == "" {
		return apperror.NewBadRequest("verification token is required")
	}

	// Get token from Token table
	tokenModel, err := s.tokenRepo.GetTokenByValue(token, model.TokenTypeEmailVerification)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if tokenModel == nil {
		return apperror.NewBadRequest("invalid verification token")
	}

	// Check if token is expired
	if tokenModel.ExpireAt.Before(time.Now()) {
		return apperror.NewBadRequest("verification token has expired")
	}

	// Get user
	user, _, err := s.userRepo.GetUserByID(tokenModel.UserID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Check if already verified
	if user.IsEmailVerified {
		return apperror.NewBadRequest("email is already verified")
	}

	// Update user email verification status
	if err := s.userRepo.UpdateEmailVerified(user.UserID, true); err != nil {
		return apperror.NewInternal(err)
	}

	// Revoke the token
	if err := s.tokenRepo.RevokeToken(tokenModel.TokenID); err != nil {
		// Log error but don't fail verification
		fmt.Printf("Warning: failed to revoke token: %v\n", err)
	}

	return nil
}

// ResendVerificationEmail resends verification email to user
func (s *userServiceImpl) ResendVerificationEmail(email string) error {
	if email == "" {
		return apperror.NewBadRequest("email is required")
	}

	// Get user by email
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("user with email '%s' not found", email)
	}

	// Check if already verified
	if user.IsEmailVerified {
		return apperror.NewBadRequest("email is already verified")
	}

	// Delete old verification tokens for this user
	if err := s.tokenRepo.DeleteTokensByUserAndType(user.UserID, model.TokenTypeEmailVerification); err != nil {
		// Log error but continue
		fmt.Printf("Warning: failed to delete old tokens: %v\n", err)
	}

	// Generate new verification token
	verificationToken, err := GenerateVerificationToken()
	if err != nil {
		return apperror.NewInternal(fmt.Errorf("failed to generate verification token: %w", err))
	}

	// Save new token to Token table
	tokenExpiry := GetVerificationTokenExpiry()
	tokenModel := &model.Token{
		UserID:    user.UserID,
		Token:     verificationToken,
		TokenType: model.TokenTypeEmailVerification,
		IsRevoked: false,
		ExpireAt:  tokenExpiry,
	}
	if err := s.tokenRepo.CreateToken(tokenModel); err != nil {
		return apperror.NewInternal(fmt.Errorf("failed to save verification token: %w", err))
	}

	// Send verification email
	if s.emailService != nil {
		if err := s.emailService.SendVerificationEmail(user.Email, verificationToken); err != nil {
			return apperror.NewInternal(fmt.Errorf("failed to send verification email: %w", err))
		}
	} else {
		return apperror.NewInternal(fmt.Errorf("email service is not configured"))
	}

	return nil
}
