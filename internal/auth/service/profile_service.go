package service

import (
	"context"
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

// UpdateProfile updates user profile information
func (s *userServiceImpl) UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	// Validate that at least one profile field is being updated
	if profile.AvatarURL == "" && profile.RankName == "" && profile.Location == "" &&
		profile.Bio == "" && profile.Level == 0 && profile.XP == 0 &&
		profile.LearningStreak == 0 && profile.LearningCount == 0 && profile.HourLearned == 0 {
		return apperror.NewBadRequest("no profile fields to update")
	}

	// Check if user exists
	user, _, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}

	// Update profile in repository
	if err := s.userRepo.UpdateProfile(ctx, userID, profile); err != nil {
		s.logger.ErrorContext(ctx, "update profile failed", "error", err, "user_id", userID)
		return apperror.NewInternal("failed to update profile: %w", err)
	}

	s.logger.InfoContext(ctx, "profile updated successfully", "user_id", userID)
	return nil
}
