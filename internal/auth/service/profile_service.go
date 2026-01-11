package service

import (
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

// UpdateProfile updates user profile information
func (s *userServiceImpl) UpdateProfile(userID string, profile *model.Profile) error {
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
	user, _, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}

	// Update profile in repository
	if err := s.userRepo.UpdateProfile(userID, profile); err != nil {
		return apperror.NewInternal(err)
	}

	return nil
}
