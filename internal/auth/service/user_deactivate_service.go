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
		days = model.DefaultDeactivationGracePeriod
	}

	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}

	until := time.Now().UTC().AddDate(0, 0, days)
	if err := s.repo.DeactivateAccountWithTokenRevoke(ctx, userID, until); err != nil {
		return apperror.NewInternal("failed to deactivate account: %w", err)
	}

	s.logger.InfoContext(ctx, "account deactivated successfully", "user_id", userID, "until", until)
	return nil
}
