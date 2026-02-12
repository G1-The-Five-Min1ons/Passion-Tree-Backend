package service

import (
	"context"
	"strings"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// Login authenticates user and returns token
// identifier can be either username or email
func (s *userServiceImpl) Login(ctx context.Context, identifier string, password string) (string, error) {
	if identifier == "" || password == "" {
		return "", apperror.NewBadRequest("identifier and password are required")
	}

	// Try to find user by email first, then by username
	var user *model.User
	var err error

	// Check if identifier is email (contains @)
	if strings.Contains(identifier, "@") {
		user, err = s.userRepo.GetUserByEmail(ctx, identifier)
	} else {
		user, err = s.userRepo.GetUserByUsername(ctx, identifier)
	}

	if err != nil || user == nil {
		s.logger.WarnContext(ctx, "login failed: user not found", "identifier", identifier)
		return "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// [Check Lock Account]
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
        timeLeft := time.Until(*user.LockedUntil).Minutes()
        s.logger.WarnContext(ctx, "login blocked: account locked", 
            "user_id", user.UserID, 
            "locked_until", user.LockedUntil)
        return "", apperror.NewTooManyRequests("account locked. try again in %.0f minutes", timeLeft)
    }

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.handleFailedLogin(ctx, user)
		return "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// Reset failed attempts on successful login
	if user.FailedAttempts > 0 {
		_ = s.userRepo.ResetFailedLogin(ctx, user.UserID)
	}

	// Generate JWT token
	jwtService := jwt.NewService()
	token, err := jwtService.GenerateAccessToken(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "jwt generation failed", "err", err, "uid", user.UserID)
		return "", apperror.NewInternal("failed to generate token: %w", err)
	}

	s.logger.InfoContext(ctx, "login success", "uid", user.UserID)
	return token, nil
}

func (s *userServiceImpl) handleFailedLogin(ctx context.Context, user *model.User) {
	newAttempts := user.FailedAttempts + 1
	var lockedUntil *time.Time

	// if failed attempts reach 5, lock account for 15 minutes
	if newAttempts >= 5 {
		t := time.Now().UTC().Add(15 * time.Minute)
		lockedUntil = &t
		s.logger.WarnContext(ctx, "account lockout triggered", "uid", user.UserID)
	}

	// Update database with new failed login attempts and lock time
	err := s.userRepo.UpdateFailedLogin(ctx, user.UserID, newAttempts, lockedUntil)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update login attempts", "error", err, "user_id", user.UserID)
	}
}

// ValidateToken validates JWT token and returns user
func (s *userServiceImpl) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, apperror.NewBadRequest("token is required")
	}

	// Validate JWT token
	jwtService := jwt.NewService()
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		return nil, apperror.NewUnauthorized("invalid token")
	}

	// Get user from database
	user, _, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "validate_token_db_failed", "err", err, "uid", claims.UserID)
		return nil, apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		s.logger.WarnContext(ctx, "token_valid_but_user_missing", "uid", claims.UserID)
		return nil, apperror.NewNotFound("user not found")
	}

	return user, nil
}
