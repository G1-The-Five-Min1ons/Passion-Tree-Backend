package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

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
		user, err = s.repo.GetUserByEmail(ctx, identifier)
	} else {
		user, err = s.repo.GetUserByUsername(ctx, identifier)
	}

	if err != nil || user == nil {
		s.logger.WarnContext(ctx, "login failed: user not found", "identifier", identifier)
		return "", "", apperror.NewUnauthorized("invalid username/email or password")
	}

	deactivatedUntil, err := s.repo.GetAccountDeactivatedUntil(ctx, user.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check account deactivation", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to verify account status")
	}
	if deactivatedUntil != nil {
		now := time.Now().UTC()
		if deactivatedUntil.After(now) {
			timeLeft := time.Until(*deactivatedUntil).Round(time.Hour)
			s.logger.WarnContext(ctx, "login blocked: account deactivated", "user_id", user.UserID, "until", deactivatedUntil)
			return "", "", apperror.NewForbidden("account is temporarily deactivated. try again in %s", timeLeft.String()).WithType("ACCOUNT_DEACTIVATED")
		}
		// Grace period expired, auto-clear deactivation
		_ = s.repo.ClearAccountDeactivatedUntil(ctx, user.UserID)
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

	// Check if 2FA is required (security flag set after token theft)
	if user.Require2FANextLogin {
		s.logger.WarnContext(ctx, "2FA verification required", "user_id", user.UserID)

		// Generate 2FA OTP
		otpCode, err := GenerateVerificationToken()
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to generate 2FA OTP", "error", err)
			return "", "", apperror.NewInternal("failed to generate security code")
		}

		// Delete old 2FA tokens for this user
		_ = s.repo.DeleteTokensByUserAndType(ctx, user.UserID, model.TokenType2FA)

		// Store 2FA OTP in database
		otpExpiry := GetVerificationTokenExpiry() // 15 minutes
		otpToken := &model.Token{
			UserID:    user.UserID,
			Token:     otpCode,
			TokenType: model.TokenType2FA,
			IsRevoked: false,
			ExpireAt:  otpExpiry,
		}
		if err := s.repo.CreateToken(ctx, otpToken); err != nil {
			s.logger.ErrorContext(ctx, "failed to store 2FA OTP", "error", err)
			return "", "", apperror.NewInternal("failed to store security code")
		}

		// Send OTP to user's email
		if s.emailService != nil {
			if err := s.emailService.SendVerificationEmail(ctx, user.Email, otpCode); err != nil {
				s.logger.ErrorContext(ctx, "failed to send 2FA OTP email (1st attempt)", "error", err, "user_id", user.UserID)

				// ถอยออกมาตั้งหลักสัก 1-2 วินาทีก่อนลองใหม่
				time.Sleep(1 * time.Second)

				if retryErr := s.emailService.SendVerificationEmail(ctx, user.Email, otpCode); retryErr != nil {
					s.logger.ErrorContext(ctx, "failed to send 2FA OTP email (2nd attempt)", "error", retryErr, "user_id", user.UserID)
					// แจ้งให้ User ทราบใน Error Message ว่าอีเมลอาจจะมาช้าหน่อย
				}
			}
		}

		s.logger.InfoContext(ctx, "2FA OTP sent to email", "user_id", user.UserID, "email", user.Email)
		return "", "", apperror.NewForbidden("security verification required. we've sent a verification code to your email: %s", user.Email)
	}

	// Reset failed attempts on successful login
	if user.FailedAttempts > 0 {
		_ = s.repo.ResetFailedLogin(ctx, user.UserID)
	}

	otpCode, err := GenerateVerificationToken()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate otp", "error", err)
		return "", "", apperror.NewInternal("failed to generate verification code")
	}

	_ = s.repo.DeleteTokensByUserAndType(ctx, user.UserID, model.TokenTypeEmailVerification)

	otpToken := &model.Token{
		UserID:    user.UserID,
		Token:     otpCode,
		TokenType: model.TokenTypeEmailVerification,
		IsRevoked: false,
		ExpireAt:  time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.CreateToken(ctx, otpToken); err != nil {
		s.logger.ErrorContext(ctx, "failed to store otp", "error", err)
		return "", "", apperror.NewInternal("failed to store verification code")
	}

	s.logger.InfoContext(ctx, "[DEBUG] OTP stored successfully",
		"otp_code", otpCode,
		"user_id", user.UserID,
		"token_type", model.TokenTypeEmailVerification,
		"expire_at", otpToken.ExpireAt,
	)

	if err := s.emailService.SendVerificationEmail(ctx, user.Email, otpCode); err != nil {
		s.logger.ErrorContext(ctx, "failed to send otp email", "error", err)
		return "", "", apperror.NewInternal("failed to send verification email")
	}

	s.logger.InfoContext(ctx, "login step 1 success, otp sent", "user_id", user.UserID)
	return "", "", apperror.NewForbidden("verification_required: please enter the 6-digit code sent to %s", user.Email)
}

func (s *userServiceImpl) handleFailedLogin(ctx context.Context, user *model.User) {
	newAttempts, err := s.repo.UpdateFailedLogin(ctx, user.UserID, 15*time.Minute)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed_update_attempts", "error", err, "user_id", user.UserID)
		return
	}

	if newAttempts >= 5 {
		s.logger.WarnContext(ctx, "account_locked", "user_id", user.UserID, "attempts", newAttempts)
	}
}

// ValidateToken validates JWT token and returns user
func (s *userServiceImpl) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, apperror.NewBadRequest("token is required")
	}

	// Validate JWT token
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, apperror.NewUnauthorized("invalid token")
	}

	// Get user from database
	user, _, err := s.repo.GetUserByID(ctx, claims.UserID)
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
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.WarnContext(ctx, "invalid refresh token", "error", err)
		return "", "", apperror.NewUnauthorized("invalid or expired refresh token")
	}

	// Get token from database to check rotation status and absolute expiration
	storedToken, err := s.repo.GetTokenByValue(ctx, refreshToken, model.TokenTypeRefresh)
	if err != nil || storedToken == nil {
		s.logger.WarnContext(ctx, "refresh token not found in database", "error", err)
		return "", "", apperror.NewUnauthorized("invalid refresh token")
	}

	// Check if token is revoked
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
		// use Task Queue to revoke token asynchronously to avoid blocking user request in Future
		_ = s.repo.RevokeTokenByValue(context.Background(), refreshToken, model.TokenTypeRefresh)
		s.logger.WarnContext(ctx, "absolute token lifetime exceeded", "user_id", claims.UserID)
		return "", "", apperror.NewUnauthorized("session expired - please login again")
	}

	// Get user from database
	user, _, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
		s.logger.ErrorContext(ctx, "user not found for refresh token", "user_id", claims.UserID)
		return "", "", apperror.NewUnauthorized("invalid refresh token")
	}

	// Generate new access token
	newAccessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate access token", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to generate access token")
	}

	// Generate new refresh token (token rotation)
	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate refresh token", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to generate refresh token")
	}

	// Mark old token as rotated (not revoked yet - for grace period)
	err = s.repo.MarkTokenAsRotated(ctx, refreshToken, model.TokenTypeRefresh)
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
	err = s.repo.CreateToken(ctx, newTokenModel)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to store new refresh token", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to store refresh token: %w", err)
	}

	// Revoke old token after short grace period (5 minutes) to handle race conditions
	// Use context.Background() because the parent request context will be cancelled before this completes
	go func() {
		time.Sleep(5 * time.Minute)
		_ = s.repo.RevokeTokenByValue(context.Background(), refreshToken, model.TokenTypeRefresh)
	}()

	s.logger.InfoContext(ctx, "tokens refreshed successfully", "user_id", user.UserID)
	return newAccessToken, newRefreshToken, nil
}

// handleTokenTheft handles potential token theft by revoking all refresh tokens
func (s *userServiceImpl) handleTokenTheft(ctx context.Context, userID string) {
	// Revoke ALL refresh tokens for this user
	err := s.repo.RevokeAllUserTokens(ctx, userID, model.TokenTypeRefresh)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to revoke tokens after theft detection", "error", err, "user_id", userID)
	}

	s.logger.WarnContext(ctx, "SECURITY: All sessions revoked due to token theft detection", "user_id", userID)

	// Set require_2fa_next_login flag to true
	err = s.repo.SetRequire2FANextLogin(ctx, userID, true)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to set 2FA requirement flag", "error", err, "user_id", userID)
	} else {
		s.logger.InfoContext(ctx, "2FA required on next login", "user_id", userID)
	}

	// Send security alert email to user
	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err == nil && user != nil && s.emailService != nil {
		if err := s.emailService.SendSecurityAlertEmail(ctx, user.Email, userID); err != nil {
			s.logger.ErrorContext(ctx, "failed to send security alert email", "error", err, "user_id", userID)
		} else {
			s.logger.InfoContext(ctx, "security alert email sent", "user_id", userID, "email", user.Email)
		}
	} else {
		s.logger.WarnContext(ctx, "security alert email not sent", "user_id", userID, "reason", "user not found or email service not configured")
	}

	s.logger.InfoContext(ctx, "SECURITY: User must verify identity (2FA) on next login", "user_id", userID)
}

// Logout revokes all refresh tokens for a user
func (s *userServiceImpl) Logout(ctx context.Context, userID string) error {
	if userID == "" {
		return apperror.NewBadRequest("user ID is required")
	}

	// Revoke all refresh tokens for the user
	err := s.repo.RevokeAllUserTokens(ctx, userID, model.TokenTypeRefresh)
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
	tokens, err := s.repo.GetActiveUserSessions(ctx, userID, model.TokenTypeRefresh)
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
	err := s.repo.RevokeTokenByIDForUser(ctx, sessionID, userID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to revoke session", "error", err, "user_id", userID, "session_id", sessionID)
		return apperror.NewNotFound("session not found")
	}

	s.logger.InfoContext(ctx, "session revoked", "user_id", userID, "session_id", sessionID)
	return nil
}

// ReactivateAccount clears the deactivation and allows account to be used again
func (s *userServiceImpl) ReactivateAccount(ctx context.Context, userID string) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}

	// Check if account is actually deactivated
	deactivatedUntil, err := s.repo.GetAccountDeactivatedUntil(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check account deactivation", "error", err, "user_id", userID)
		return apperror.NewInternal("failed to verify account status: %w", err)
	}

	if deactivatedUntil == nil {
		return apperror.NewBadRequest("account is not deactivated")
	}

	// Clear the deactivation
	if err := s.repo.ClearAccountDeactivatedUntil(ctx, userID); err != nil {
		s.logger.ErrorContext(ctx, "failed to clear account deactivation", "error", err, "user_id", userID)
		return apperror.NewInternal("failed to reactivate account: %w", err)
	}

	s.logger.InfoContext(ctx, "account reactivated successfully", "user_id", userID)
	return nil
}

// GetAccountDeactivatedUntil returns the deactivation timestamp if account is deactivated, nil otherwise
func (s *userServiceImpl) GetAccountDeactivatedUntil(ctx context.Context, userID string) (*time.Time, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	deactivatedUntil, err := s.repo.GetAccountDeactivatedUntil(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get account deactivation status", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to check account status: %w", err)
	}

	return deactivatedUntil, nil
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil || *s == "" {
		return "Unknown Device"
	}
	return *s
}
