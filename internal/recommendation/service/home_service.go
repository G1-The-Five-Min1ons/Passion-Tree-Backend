package service

import (
	"context"
	"passiontree/internal/pkg/apperror"
	recmodel "passiontree/internal/recommendation/model"
)

func (s *serviceImpl) RecommendHomePathsForUser(ctx context.Context, userID string) (*recmodel.RecommendPathResponse, error) {
	if userID == "" {
		return nil, apperror.NewUnauthorized("Authentication session expired")
	}

	s.logger.InfoContext(ctx, "fetching saved home recommendations", "user_id", userID)

	savedPaths, err := s.recRepo.GetSavedHomeRecommendations(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch saved home recommendations", "error", err.Error())
		return nil, apperror.NewInternal("could not retrieve your recommendations")
	}

	if len(savedPaths) > 0 {
		return &recmodel.RecommendPathResponse{
			UserPersonaQuery: "Personalized recommendations ready.",
			RecommendedPaths: savedPaths,
		}, nil
	}

	s.logger.InfoContext(ctx, "no saved recommendations found, using fallback popular paths", "user_id", userID)
	return s.getFallbackPopularPaths(ctx, "Welcome! Here are some popular learning paths to get you started.")
}

func (s *serviceImpl) getFallbackPopularPaths(ctx context.Context, message string) (*recmodel.RecommendPathResponse, error) {
	topPaths, err := s.recRepo.GetTopPopularPaths(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch fallback popular paths", "error", err.Error())
		return nil, apperror.NewInternal("failed to fetch popular paths")
	}

	return &recmodel.RecommendPathResponse{
		UserPersonaQuery: message,
		RecommendedPaths: topPaths,
	}, nil
}
