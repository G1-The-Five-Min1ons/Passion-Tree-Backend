package service

import (
	"context"
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

func (s *userServiceImpl) CreateUserByAdmin(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	if user == nil {
		return "", apperror.NewBadRequest("user payload is required")
	}
	if profile == nil {
		profile = &model.Profile{}
	}

	return s.CreateUser(ctx, user, profile)
}

// GetAllUsers retrieves all users (admin only)
func (s *userServiceImpl) GetAllUsers(ctx context.Context) ([]*model.UserWithProfile, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get all users", "error", err)
		return nil, apperror.NewInternal("failed to retrieve users: %w", err)
	}
	return users, nil
}

// GetDashboardStats retrieves dashboard statistics (admin only)
func (s *userServiceImpl) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	stats, err := s.repo.GetDashboardStats(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get dashboard stats", "error", err)
		return nil, apperror.NewInternal("failed to retrieve dashboard stats: %w", err)
	}
	return stats, nil
}

func (s *userServiceImpl) UpdateUserByAdmin(ctx context.Context, id string, firstName string, lastName string, role string) error {
	if role != "" && role != string(model.RoleStudent) && role != string(model.RoleTeacher) && role != string(model.RoleAdmin) && role != string(model.RolePending) && role != "user" && role != "moderator" {
		return apperror.NewBadRequest("role must be one of 'student', 'teacher', 'admin', 'pending', 'user', or 'moderator'")
	}

	return s.UpdateUser(ctx, id, "", firstName, lastName, role)
}

func (s *userServiceImpl) DeleteUserByAdmin(ctx context.Context, id string) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	user, _, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user not found")
	}

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return apperror.NewInternal("failed to delete user: %w", err)
	}

	s.logger.InfoContext(ctx, "user deleted by admin", "user_id", id)
	return nil
}
