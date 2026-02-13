package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// Login authenticates user and returns access and refresh tokens
// identifier can be either username or email
func (s *userServiceImpl) Login(ctx context.Context, identifier string, password string, deviceInfo, ipAddress, userAgent string) (string, string, error) {
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

	// Get absolute expiration time from config
	refreshTTLHours := 168 // 7 days default
	refreshAbsoluteHours := 720 // 30 days default
	if s.config.JWTRefreshTTL != "" {
		if hours, parseErr := strconv.Atoi(s.config.JWTRefreshTTL); parseErr == nil {
			refreshTTLHours = hours
		}
	}
	if s.config.JWTRefreshAbsolute != "" {
		if hours, parseErr := strconv.Atoi(s.config.JWTRefreshAbsolute); parseErr == nil {
			refreshAbsoluteHours = hours
		}
	}

	// Store refresh token in database with session tracking
	now := time.Now()
	expireAt := now.Add(time.Duration(refreshTTLHours) * time.Hour)
	maxExpireAt := now.Add(time.Duration(refreshAbsoluteHours) * time.Hour)
	
	tokenModel := &model.Token{
		UserID:       user.UserID,
		Token:        refreshToken,
		TokenType:    model.TokenTypeRefresh,
		IsRevoked:    false,
		ExpireAt:     expireAt,
		DeviceInfo:   &deviceInfo,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		LastUsedAt:   &now,
		MaxExpiresAt: &maxExpireAt,
		IsRotated:    false,
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

// RefreshAccessToken generates new access and refresh tokens using a valid refresh token
// This implements token rotation for enhanced security
func (s *userServiceImpl) RefreshAccessToken(ctx context.Context, refreshToken string, deviceInfo, ipAddress, userAgent string) (string, string, error) {
	if refreshToken == "" {
		return "", "", apperror.NewBadRequest("refresh token is required")
	}

	// Validate refresh token
	jwtService := jwt.NewService()
	claims, err := jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.WarnContext(ctx, "invalid refresh token", "error", err)
		return "", "", apperror.NewUnauthorized("invalid or expired refresh token")
	}

	// Get token from database to check rotation status and absolute expiration
	storedToken, err := s.tokenRepo.GetTokenByValue(refreshToken, model.TokenTypeRefresh)
	if err != nil || storedToken == nil {
		s.logger.WarnContext(ctx, "refresh token not found in database", "error", err)
		return "", "", apperror.NewUnauthorized("invalid refresh token")
	}

	// 🚨 Check if token is revoked
	if storedToken.IsRevoked {
		// Token was revoked but someone is trying to use it = potential theft!
		s.logger.ErrorContext(ctx, "revoked token used - potential theft", "user_id", claims.UserID)
		s.handleTokenTheft(ctx, claims.UserID)
		return "", "", apperror.NewUnauthorized("token has been revoked - all sessions terminated for security")
	}

	// 🚨 Check if token was already rotated (reuse detection)
	if storedToken.IsRotated {
		// Old token being reused = definitely theft!
		s.logger.ErrorContext(ctx, "token reuse detected - security breach", "user_id", claims.UserID)
		s.handleTokenTheft(ctx, claims.UserID)
		return "", "", apperror.NewUnauthorized("token reuse detected - all sessions terminated for security")
	}

	// Check absolute expiration
	if storedToken.MaxExpiresAt != nil && time.Now().After(*storedToken.MaxExpiresAt) {
		_ = s.tokenRepo.RevokeTokenByValue(refreshToken, model.TokenTypeRefresh)
		s.logger.WarnContext(ctx, "absolute token lifetime exceeded", "user_id", claims.UserID)
		return "", "", apperror.NewUnauthorized("session expired - please login again")
	}

	// Get user from database
	user, _, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
		s.logger.ErrorContext(ctx, "user not found for refresh token", "user_id", claims.UserID)
		return "", "", apperror.NewUnauthorized("invalid refresh token")
	}

	// Generate new access token
	newAccessToken, err := jwtService.GenerateAccessToken(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate access token", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to generate access token")
	}

	// Generate new refresh token (token rotation)
	newRefreshToken, err := jwtService.GenerateRefreshToken(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate refresh token", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to generate refresh token")
	}

	// Mark old token as rotated (not revoked yet - for grace period)
	err = s.tokenRepo.MarkTokenAsRotated(refreshToken, model.TokenTypeRefresh)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to mark token as rotated", "error", err)
		// Continue anyway
	}

	// Get refresh TTL from config
	refreshTTLHours := 168 // 7 days default
	if s.config.JWTRefreshTTL != "" {
		if hours, parseErr := strconv.Atoi(s.config.JWTRefreshTTL); parseErr == nil {
			refreshTTLHours = hours
		}
	}

	// Store new refresh token with same absolute expiration as original
	now := time.Now()
	expireAt := now.Add(time.Duration(refreshTTLHours) * time.Hour)
	
	newTokenModel := &model.Token{
		UserID:        user.UserID,
		Token:         newRefreshToken,
		TokenType:     model.TokenTypeRefresh,
		IsRevoked:     false,
		ExpireAt:      expireAt,
		DeviceInfo:    &deviceInfo,
		IPAddress:     &ipAddress,
		UserAgent:     &userAgent,
		LastUsedAt:    &now,
		MaxExpiresAt:  storedToken.MaxExpiresAt, // Keep same absolute expiration
		ParentTokenID: &storedToken.TokenID,     // Track rotation chain
		IsRotated:     false,
	}
	err = s.tokenRepo.CreateToken(newTokenModel)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to store new refresh token", "error", err, "user_id", user.UserID)
		// Continue anyway - return the tokens
	}

	// Revoke old token after short grace period (5 minutes) to handle race conditions
	go func() {
		time.Sleep(5 * time.Minute)
		_ = s.tokenRepo.RevokeTokenByValue(refreshToken, model.TokenTypeRefresh)
	}()

	s.logger.InfoContext(ctx, "tokens refreshed successfully", "user_id", user.UserID)
	return newAccessToken, newRefreshToken, nil
}

// handleTokenTheft handles potential token theft by revoking all refresh tokens
func (s *userServiceImpl) handleTokenTheft(ctx context.Context, userID string) {
	// Revoke ALL refresh tokens for this user
	err := s.tokenRepo.RevokeAllUserTokens(userID, model.TokenTypeRefresh)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to revoke tokens after theft detection", "error", err, "user_id", userID)
	}

	s.logger.WarnContext(ctx, "SECURITY: All sessions revoked due to token theft detection", "user_id", userID)

	// TODO: Send security alert email to user
	// TODO: Consider requiring 2FA on next login
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

// GetActiveSessions retrieves all active sessions/devices for a user
func (s *userServiceImpl) GetActiveSessions(ctx context.Context, userID string, currentRefreshToken string) (*model.GetActiveSessionsResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user ID is required")
	}

	// Get all active refresh tokens for the user
	tokens, err := s.tokenRepo.GetActiveUserSessions(userID, model.TokenTypeRefresh)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get active sessions", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to retrieve sessions")
	}

	// Convert tokens to session info
	sessions := make([]*model.SessionInfo, 0, len(tokens))
	for _, token := range tokens {
		session := &model.SessionInfo{
			SessionID:  token.TokenID,
			DeviceInfo: getStringValue(token.DeviceInfo),
			IPAddress:  getStringValue(token.IPAddress),
			UserAgent:  getStringValue(token.UserAgent),
			LastUsedAt: token.LastUsedAt,
			CreatedAt:  token.CreatedAt,
			ExpiresAt:  token.ExpireAt,
			IsCurrent:  token.Token == currentRefreshToken,
		}
		sessions = append(sessions, session)
	}

	response := &model.GetActiveSessionsResponse{
		TotalSessions: len(sessions),
		Sessions:      sessions,
	}

	s.logger.InfoContext(ctx, "retrieved active sessions", "user_id", userID, "count", len(sessions))
	return response, nil
}

// LogoutSession revokes a specific session by session ID
func (s *userServiceImpl) LogoutSession(ctx context.Context, userID string, sessionID string) error {
	if userID == "" || sessionID == "" {
		return apperror.NewBadRequest("user ID and session ID are required")
	}

	// Revoke the token only if it belongs to the user (security check in SQL)
	err := s.tokenRepo.RevokeTokenByIDForUser(sessionID, userID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to revoke session", "error", err, "user_id", userID, "session_id", sessionID)
		return apperror.NewNotFound("session not found")
	}

	s.logger.InfoContext(ctx, "session revoked", "user_id", userID, "session_id", sessionID)
	return nil
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return "Unknown"
	}
	return *s
}

