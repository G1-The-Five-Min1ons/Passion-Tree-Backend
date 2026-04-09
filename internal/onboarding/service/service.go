package service

import (
	"context"
	"log/slog"

	"passiontree/internal/onboarding/model"
	"passiontree/internal/onboarding/repository"
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
