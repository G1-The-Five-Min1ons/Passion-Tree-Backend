package service

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

// VerifyEmail verifies a user's email
func (s *userServiceImpl) VerifyEmail(ctx context.Context, otpCode string, deviceInfo, ip, ua string) (string, string, error) {
	if otpCode == "" {
		s.logger.ErrorContext(ctx, "verification code is empty")
		return "", "", apperror.NewBadRequest("verification code is required")
	}

	storedToken, err := s.repo.GetTokenByValue(ctx, otpCode, model.TokenTypeEmailVerification)
	s.logger.InfoContext(ctx, "[DEBUG] GetTokenByValue result",
		"code", otpCode,
		"token_type", model.TokenTypeEmailVerification,
		"found", storedToken != nil,
		"db_error", err,
	)
	if err != nil || storedToken == nil {
		s.logger.WarnContext(ctx, "invalid otp code", "code", otpCode, "error", err)
		return "", "", apperror.NewBadRequest("invalid or expired verification code")
	}

	if storedToken.ExpireAt.Before(time.Now()) {
		s.logger.WarnContext(ctx, "verification code expired", "code", otpCode, "expire_at", storedToken.ExpireAt)
		return "", "", apperror.NewBadRequest("verification code has expired")
	}

	user, _, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil || user == nil {
		s.logger.ErrorContext(ctx, "user not found for verification code", "error", err, "user_id", storedToken.UserID)
		return "", "", apperror.NewNotFound("user not found")
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		s.logger.WarnContext(ctx, "verify email failed: account locked", "user_id", user.UserID)
		return "", "", apperror.NewTooManyRequests("account is currently locked")
	}

	if err := s.repo.UpdateEmailVerified(ctx, user.UserID, true); err != nil {
		s.logger.ErrorContext(ctx, "failed to update email verified status", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to update verification status")
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate tokens after email verification", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to generate session")
	}

	now := time.Now()
	tokenModel := &model.Token{
		UserID:     user.UserID,
		Token:      refreshToken,
		TokenType:  model.TokenTypeRefresh,
		ExpireAt:   now.Add(168 * time.Hour), // 7 วัน
		DeviceInfo: &deviceInfo,
		IPAddress:  &ip,
		UserAgent:  &ua,
	}

	if err := s.repo.CreateToken(ctx, tokenModel); err != nil {
		s.logger.ErrorContext(ctx, "failed to store refresh token", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to create session")
	}

	if err := s.repo.RevokeTokenByValue(ctx, otpCode, model.TokenTypeEmailVerification); err != nil {
		s.logger.WarnContext(ctx, "failed to revoke verification token", "error", err, "user_id", user.UserID)
		// Non-critical: token will expire naturally, continue
	}

	s.logger.InfoContext(ctx, "otp verified and tokens generated", "user_id", user.UserID)
	return accessToken, refreshToken, nil
}

// ResendVerificationEmail resends verification email to user
func (s *userServiceImpl) ResendVerificationEmail(ctx context.Context, email string) error {
	if email == "" {
		return apperror.NewBadRequest("email is required")
	}

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		s.logger.WarnContext(ctx, "resend verification email failed: user not found", "email", email)
		return apperror.NewBadRequest("user with this email does not exist")
	}

	if user.IsEmailVerified {
		s.logger.InfoContext(ctx, "resending login verification otp for verified user", "user_id", user.UserID, "email", email)
	} else {
		s.logger.InfoContext(ctx, "resending account verification otp for unverified user", "user_id", user.UserID, "email", email)
	}

	_ = s.repo.DeleteTokensByUserAndType(ctx, user.UserID, model.TokenTypeEmailVerification)
	otpCode, _ := GenerateVerificationToken()

	otpToken := &model.Token{
		UserID:    user.UserID,
		Token:     otpCode,
		TokenType: model.TokenTypeEmailVerification,
		IsRevoked: false,
		ExpireAt:  GetVerificationTokenExpiry(),
	}

	if err := s.repo.CreateToken(ctx, otpToken); err != nil {
		s.logger.ErrorContext(ctx, "failed to store verification code", "error", err, "user_id", user.UserID)
		return apperror.NewInternal("failed to store verification code")
	}

	// Send verification email with JWT token
	if err := s.emailService.SendVerificationEmail(ctx, user.Email, otpCode); err != nil {
		s.logger.ErrorContext(ctx, "failed to send verification email", "error", err, "email", user.Email)
		return apperror.NewInternal("failed to send email")
	}

	s.logger.InfoContext(ctx, "verification otp sent", "user_id", user.UserID)
	return nil
}
