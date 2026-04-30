package service

import (
	"context"
	"time"

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

	s.recomputeRecommendationsAsync(userID)

	s.logger.InfoContext(ctx, "onboarding saved", "user_id", userID)
	return nil
}

// recomputeRecommendationsAsync triggers a personalized recommendation
// recompute for the user without blocking the onboarding response. Failures
// are logged; the daily batch cron remains the safety net.
func (s *serviceImpl) recomputeRecommendationsAsync(userID string) {
	if s.recomputer == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		if err := s.recomputer.RecomputeForUser(ctx, userID); err != nil {
			s.logger.ErrorContext(ctx, "failed to recompute recommendations after onboarding", "user_id", userID, "error", err)
		}
	}()
}
