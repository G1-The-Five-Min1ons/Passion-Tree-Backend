package service

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

// VerifyEmail verifies a user's email using verification token
func (s *serviceImpl) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return apperror.NewBadRequest("verification token is required")
	}

	// Get token from Token table
	tokenModel, err := s.repo.GetTokenByValue(ctx, token, model.TokenTypeEmailVerification)
	if err != nil {
		s.logger.ErrorContext(ctx, "verify email token failed", "error", err)
		return apperror.NewInternal("failed to get verification token: %w", err)
	}
	if tokenModel == nil {
		return apperror.NewBadRequest("invalid or expired verification token")
	}

	// Check if token is expired
	if tokenModel.ExpireAt.Before(time.Now()) {
		return apperror.NewBadRequest("verification token has expired")
	}

	// Get user
	user, _, err := s.repo.GetUserByID(ctx, tokenModel.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "verify email get user id failed", "error", err, "user_id", tokenModel.UserID)
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Check if already verified
	if user.IsEmailVerified {
		return apperror.NewBadRequest("email already verified")
	}

	// Verify email and revoke token in transaction
	if err := s.repo.VerifyEmailAndRevokeToken(ctx, tokenModel.UserID, tokenModel.TokenID); err != nil {
		s.logger.ErrorContext(ctx, "verify email transaction failed", "error", err, "user_id", tokenModel.UserID)
		return apperror.NewInternal("failed to verify email: %w", err)
	}

	s.logger.InfoContext(ctx, "email verified successfully", "user_id", tokenModel.UserID)
	return nil
}

// ResendVerificationEmail resends verification email to user
func (s *serviceImpl) ResendVerificationEmail(ctx context.Context, email string) error {
	if email == "" {
		return apperror.NewBadRequest("email is required")
	}

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
    if err != nil || user == nil {
        return apperror.NewNotFound("user not found")
    }

    if user.IsEmailVerified {
        return apperror.NewBadRequest("email already verified")
    }

	// Delete old verification tokens for this user
	if err := s.repo.DeleteTokensByUserAndType(ctx, user.UserID, model.TokenTypeEmailVerification); err != nil {
        s.logger.WarnContext(ctx, "resend cleanup failed", "error", err, "user_id", user.UserID)
    }

	// Generate new verification token
	verificationToken, err := GenerateVerificationToken()
	if err != nil {
		s.logger.ErrorContext(ctx, "generate verification token failed", "error", err)
		return apperror.NewInternal("failed to generate verification token: %w", err)
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
	if err := s.repo.CreateToken(ctx, tokenModel); err != nil {
		s.logger.ErrorContext(ctx, "save verification token failed", "error", err, "user_id", user.UserID)
		return apperror.NewInternal("failed to save verification token: %w", err)
	}

	// Send verification email
	if err := s.SendVerificationEmail(user.Email, verificationToken); err != nil {
		s.logger.ErrorContext(ctx, "send verification email failed", "error", err, "email", user.Email)
		return apperror.NewInternal("failed to send verification email: %w", err)
	}

	s.logger.InfoContext(ctx, "verification email resent successfully", "user_id", user.UserID)
	return nil
}
