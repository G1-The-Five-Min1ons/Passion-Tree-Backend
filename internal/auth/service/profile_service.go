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

	// Check if user exists
	user, _, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}

	// TODO: Implement profile update in repository
	// For now, return not implemented
	return apperror.NewInternal(nil)
}