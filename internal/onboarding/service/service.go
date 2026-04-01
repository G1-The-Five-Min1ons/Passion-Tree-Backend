package service

import (
	"context"
	"log/slog"

	"passiontree/internal/onboarding/model"
	"passiontree/internal/onboarding/repository"
	"passiontree/internal/pkg/apperror"
)

type Service interface {
	SaveOnboarding(ctx context.Context, userID string, req model.SaveOnboardingRequest) error
	GetOnboarding(ctx context.Context, userID string) (*model.OnboardingData, error)
}

type serviceImpl struct {
	repo   repository.Repository
	logger *slog.Logger
}

func NewService(repo repository.Repository, logger *slog.Logger) Service {
	return &serviceImpl{repo: repo, logger: logger}
}

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

func (s *serviceImpl) GetOnboarding(ctx context.Context, userID string) (*model.OnboardingData, error) {
	data, err := s.repo.GetOnboardingByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("failed to get onboarding: %v", err)
	}
	return data, nil
}
