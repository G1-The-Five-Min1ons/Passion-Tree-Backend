package service

import (
	"context"

	"passiontree/internal/onboarding/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) SaveOnboarding(ctx context.Context, userID string, req model.SaveOnboardingRequest) error {
	if len(req.Subjects) == 0 {
		return apperror.NewBadRequest("subjects is required")
	}
	if req.KnowledgeLevel == "" {
		return apperror.NewBadRequest("knowledge_level is required")
	}
	if req.Motivation == "" {
		return apperror.NewBadRequest("motivation is required")
	}
	if req.DailyGoal == "" {
		return apperror.NewBadRequest("daily_goal is required")
	}
	if len(req.LearningStyles) == 0 {
		return apperror.NewBadRequest("learning_styles is required")
	}
	if req.ReflectionHabit == "" {
		return apperror.NewBadRequest("reflection_habit is required")
	}

	if err := s.repo.UpsertOnboarding(ctx, userID, req); err != nil {
		return apperror.NewInternal("failed to save onboarding: %v", err)
	}

	s.logger.InfoContext(ctx, "onboarding saved", "user_id", userID)
	return nil
}
