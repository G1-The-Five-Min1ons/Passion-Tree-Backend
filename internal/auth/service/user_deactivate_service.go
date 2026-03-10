package service

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

func (s *userServiceImpl) DeactivateAccount(ctx context.Context, userID string, days int) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}
	if days <= 0 {
		days = 14
	}

	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}

	until := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
	if err := s.repo.SetAccountDeactivatedUntil(ctx, userID, until); err != nil {
		return apperror.NewInternal("failed to deactivate account: %w", err)
	}

	if err := s.repo.RevokeAllUserTokens(ctx, userID, model.TokenTypeRefresh); err != nil {
		return apperror.NewInternal("failed to revoke sessions: %w", err)
	}

	s.logger.InfoContext(ctx, "account deactivated successfully", "user_id", userID, "until", until)
	return nil
}
