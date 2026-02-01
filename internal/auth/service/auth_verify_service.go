package service

import (
	"context"
	"fmt"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

// VerifyEmail verifies a user's email using verification token
func (s *userServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return apperror.NewBadRequest("verification token is required")
	}

	// Get token from Token table
	tokenModel, err := s.tokenRepo.GetTokenByValue(token, model.TokenTypeEmailVerification)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if tokenModel == nil {
		return apperror.NewBadRequest("invalid or expired verification token")
	}

	// Check if token is expired
	if tokenModel.ExpireAt.Before(time.Now()) {
		return apperror.NewBadRequest("verification token has expired")
	}

	// Get user
	user, _, err := s.userRepo.GetUserByID(ctx, tokenModel.UserID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Check if already verified
	if user.IsEmailVerified {
		return apperror.NewBadRequest("email already verified")
	}

	// Update user email verification status
	if err := s.userRepo.UpdateEmailVerified(ctx, tokenModel.UserID, true); err != nil {
		return apperror.NewInternal(err)
	}

	// Revoke the token
	if err := s.tokenRepo.RevokeToken(tokenModel.TokenID); err != nil {
		// Log error but don't fail verification
		fmt.Printf("Warning: failed to revoke verification token: %v\n", err)
	}

	return nil
}

// ResendVerificationEmail resends verification email to user
func (s *userServiceImpl) ResendVerificationEmail(ctx context.Context, email string) error {
	if email == "" {
		return apperror.NewBadRequest("email is required")
	}

	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	// Check if already verified
	if user.IsEmailVerified {
		return apperror.NewBadRequest("email already verified")
	}

	// Delete old verification tokens for this user
	if err := s.tokenRepo.DeleteTokensByUserAndType(user.UserID, model.TokenTypeEmailVerification); err != nil {
		// Log error but continue
		fmt.Printf("Warning: failed to delete old verification tokens: %v\n", err)
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
