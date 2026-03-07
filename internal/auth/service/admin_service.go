package service

import (
	"context"
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

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
