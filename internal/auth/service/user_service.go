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

// Login authenticates user and returns access and refresh tokens
// identifier can be either username or email
func (s *userServiceImpl) Login(ctx context.Context, identifier string, password string) (string, string, error) {
	if identifier == "" || password == "" {
		return "", "", apperror.NewBadRequest("identifier and password are required")
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
		return "", "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// [Check Lock Account]
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
        timeLeft := time.Until(*user.LockedUntil).Minutes()
        s.logger.WarnContext(ctx, "login blocked: account locked", 
            "user_id", user.UserID, 
            "locked_until", user.LockedUntil)
        return "", "", apperror.NewTooManyRequests("account locked. try again in %.0f minutes", timeLeft)
    }

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.handleFailedLogin(ctx, user)
		return "", "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// Reset failed attempts on successful login
	if user.FailedAttempts > 0 {
		_ = s.userRepo.ResetFailedLogin(ctx, user.UserID)
	}

	// Generate JWT token pair
	jwtService := jwt.NewService()
	accessToken, refreshToken, err := jwtService.GenerateTokenPair(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "jwt generation failed", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to generate token: %w", err)
	}

	// Store refresh token in database
	tokenModel := &model.Token{
		UserID:    user.UserID,
		Token:     refreshToken,
		TokenType: model.TokenTypeRefresh,
		IsRevoked: false,
		ExpireAt:  time.Now().Add(7 * 24 * time.Hour), // 7 days
	}
	err = s.tokenRepo.CreateToken(tokenModel)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to store refresh token", "error", err, "user_id", user.UserID)
		// Continue anyway - we can still return the tokens
	}

	s.logger.InfoContext(ctx, "login success", "user_id", user.UserID)
	return accessToken, refreshToken, nil
}

func (s *userServiceImpl) handleFailedLogin(ctx context.Context, user *model.User) {
    newAttempts, err := s.userRepo.UpdateFailedLogin(ctx, user.UserID, 15*time.Minute)
    if err != nil {
        s.logger.ErrorContext(ctx, "failed_update_attempts", "error", err, "user_id", user.UserID)
        return
    }

    if newAttempts >= 5 {
        s.logger.WarnContext(ctx, "account_locked", "user_id", user.UserID, "attempts", newAttempts, )
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
		s.logger.ErrorContext(ctx, "validate_token_db_failed", "error", err, "user_id", claims.UserID)
		return nil, apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		s.logger.WarnContext(ctx, "token_valid_but_user_missing", "user_id", claims.UserID)
		return nil, apperror.NewNotFound("user not found")
	}

	return user, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (s *userServiceImpl) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", apperror.NewBadRequest("refresh token is required")
	}

	// Validate refresh token
	jwtService := jwt.NewService()
	claims, err := jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.WarnContext(ctx, "invalid refresh token", "error", err)
		return "", apperror.NewUnauthorized("invalid or expired refresh token")
	}

	// Check if token is revoked in database
	isRevoked, err := s.tokenRepo.IsTokenRevoked(refreshToken, model.TokenTypeRefresh)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check token revocation", "error", err)
		return "", apperror.NewInternal("failed to validate refresh token")
	}

	if isRevoked {
		s.logger.WarnContext(ctx, "revoked refresh token used", "user_id", claims.UserID)
		return "", apperror.NewUnauthorized("refresh token has been revoked")
	}

	// Get user from database
	user, _, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
		s.logger.ErrorContext(ctx, "user not found for refresh token", "user_id", claims.UserID)
		return "", apperror.NewUnauthorized("invalid refresh token")
	}

	// Generate new access token
	newAccessToken, err := jwtService.GenerateAccessToken(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate access token", "error", err, "user_id", user.UserID)
		return "", apperror.NewInternal("failed to generate access token")
	}

	s.logger.InfoContext(ctx, "access token refreshed", "user_id", user.UserID)
	return newAccessToken, nil
}

// Logout revokes all refresh tokens for a user
func (s *userServiceImpl) Logout(ctx context.Context, userID string) error {
	if userID == "" {
		return apperror.NewBadRequest("user ID is required")
	}

	// Revoke all refresh tokens for the user
	err := s.tokenRepo.RevokeAllUserTokens(userID, model.TokenTypeRefresh)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to revoke tokens on logout", "error", err, "user_id", userID)
		return apperror.NewInternal("failed to logout")
	}

	s.logger.InfoContext(ctx, "user logged out", "user_id", userID)
	return nil
}

