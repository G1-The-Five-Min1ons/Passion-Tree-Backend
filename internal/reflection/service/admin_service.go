package service

import (
	"context"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

// GetReflectionStats returns statistics about reflections (admin only)
func (s *serviceImpl) GetReflectionStats(ctx context.Context) (*model.ReflectionStats, error) {
	s.logger.InfoContext(ctx, "fetching reflection stats")

	stats, err := s.refRepo.GetReflectionStats(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch reflection stats", "error", err)
		return nil, apperror.NewInternal("failed to fetch reflection stats: %w", err)
	}

	s.logger.InfoContext(ctx, "successfully fetched reflection stats")
	return stats, nil
}
