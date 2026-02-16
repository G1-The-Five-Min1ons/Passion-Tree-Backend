package service

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"golang.org/x/crypto/bcrypt"
)

// ForgotPassword sends a password reset code to user's email
func (s *userServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	if email == "" {
		return apperror.NewBadRequest("email is required")
	}

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return apperror.NewInternal("failed to get user by email: %w", err)
	}
	if user == nil {
		// Don't reveal if email exists for security
		return nil
	}

	// Delete old password reset tokens for this user
	if err := s.repo.DeleteTokensByUserAndType(ctx, user.UserID, model.TokenTypePasswordReset); err != nil {
		s.logger.WarnContext(ctx, "forgot password cleanup failed", "error", err, "user_id", user.UserID)
	}

	// Generate new password reset code
	resetCode, err := GenerateVerificationToken()
	if err != nil {
		return apperror.NewInternal("failed to generate reset code: %w", err)
	}

	// Save reset code to Token table
	tokenExpiry := GetVerificationTokenExpiry()
	tokenModel := &model.Token{
		UserID:    user.UserID,
		Token:     resetCode,
		TokenType: model.TokenTypePasswordReset,
		IsRevoked: false,
		ExpireAt:  tokenExpiry,
	}
	if err := s.repo.CreateToken(ctx, tokenModel); err != nil {
		return apperror.NewInternal("failed to save reset code: %w", err)
	}

	// Send password reset email
	if s.emailService != nil {
		if err := s.emailService.SendPasswordResetEmail(user.Email, resetCode); err != nil {
			_ = s.repo.RevokeTokenByValue(ctx, tokenModel.Token, tokenModel.TokenType) // Rollback
			s.logger.ErrorContext(ctx, "forgot pwd email failed", "user_id", user.UserID, "error", err)
			return apperror.NewInternal("failed to send reset email")
		}
	}

	s.logger.InfoContext(ctx, "forgot password email sent", "user_id", user.UserID, "email", user.Email)
	return nil
}

// ResetPassword resets user password using reset code
func (s *userServiceImpl) ResetPassword(ctx context.Context, code string, newPassword string) error {
	if code == "" {
		return apperror.NewBadRequest("reset code is required")
	}
	if newPassword == "" {
		return apperror.NewBadRequest("new password is required")
	}
	if len(newPassword) < 6 {
		return apperror.NewBadRequest("password must be at least 6 characters")
	}

	// Get token from Token table
	tokenModel, err := s.repo.GetTokenByValue(ctx, code, model.TokenTypePasswordReset)
	if err != nil {
		s.logger.ErrorContext(ctx, "reset pwd get token failed", "code", code, "err", err)
		return apperror.NewInternal("failed to get token by value: %w", err)
	}
	if tokenModel == nil {
		return apperror.NewBadRequest("invalid or expired reset code")
	}

	// Check if token is expired
	if tokenModel.ExpireAt.Before(time.Now()) {
		return apperror.NewBadRequest("reset code has expired")
	}

	// Get user
	user, _, err := s.repo.GetUserByID(ctx, tokenModel.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "reset pwd get user failed", "uid", tokenModel.UserID, "err", err)
		return apperror.NewInternal("failed to get user: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorContext(ctx, "reset pwd hash failed", "user_id", user.UserID, "error", err)
		return apperror.NewInternal("failed to hash password: %w", err)
	}

	// Reset password and revoke token in transaction
	if err := s.repo.ResetPasswordWithToken(ctx, user.UserID, string(hashedPassword), tokenModel.TokenID); err != nil {
		s.logger.ErrorContext(ctx, "reset pwd failed", "user_id", user.UserID, "error", err)
		return apperror.NewInternal("failed to reset password: %w", err)
	}

	s.logger.InfoContext(ctx, "password reset successful", "user_id", user.UserID)
	return nil
}

// ChangePassword changes password for authenticated user
func (s *userServiceImpl) ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}
	if oldPassword == "" {
		return apperror.NewBadRequest("old password is required")
	}
	if newPassword == "" {
		return apperror.NewBadRequest("new password is required")
	}
	if len(newPassword) < 6 {
		return apperror.NewBadRequest("new password must be at least 6 characters")
	}

	// Get user
	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "change pwd get user failed", "user_id", userID, "error", err)
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return apperror.NewUnauthorized("invalid old password")
	}

	// Check if new password is same as old
	if oldPassword == newPassword {
		return apperror.NewBadRequest("new password must be different from old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorContext(ctx, "change pwd hash failed", "user_id", user.UserID, "error", err)
		return apperror.NewInternal("failed to hash password: %w", err)
	}

	// Update password in database
	if err := s.repo.UpdatePassword(ctx, user.UserID, string(hashedPassword)); err != nil {
		s.logger.ErrorContext(ctx, "change pwd update failed", "user_id", user.UserID, "error", err)
		return apperror.NewInternal("failed to update password: %w", err)
	}

	s.logger.InfoContext(ctx, "password change successful", "user_id", user.UserID)
	return nil
}


