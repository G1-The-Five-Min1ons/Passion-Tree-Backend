package service

import (
	"context"
	"strings"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

// UpdateProfile updates user profile information
func (s *userServiceImpl) UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	err := s.repo.UpdateProfile(ctx, userID, profile)
	
	if err != nil {
		// Check if error indicates user not found
		if strings.Contains(err.Error(), "not found") {
			s.logger.WarnContext(ctx, "update profile failed - user not found",
				"user_id", userID,
				"error", err)
			return apperror.NewNotFound("user with id '%s' not found", userID)
		}

		// Propagate other errors without wrapping to preserve original error type
		s.logger.ErrorContext(ctx, "update profile failed",
			"error", err,
			"user_id", userID)
		return err
	}

	s.logger.InfoContext(ctx, "profile updated successfully", "user_id", userID)
	return nil
}
