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

// RecommendationRecomputer is satisfied by the recommendation service. We
// declare it here so onboarding can trigger a personalized recompute right
// after a user finishes onboarding without depending on the recommendation
// package directly.
type RecommendationRecomputer interface {
	RecomputeForUser(ctx context.Context, userID string) error
}

type serviceImpl struct {
	repo       repository.Repository
	logger     *slog.Logger
	recomputer RecommendationRecomputer
}

func NewService(repo repository.Repository, logger *slog.Logger) Service {
	return &serviceImpl{repo: repo, logger: logger}
}

// SetRecommendationRecomputer injects the recommender used to seed
// personalized recommendations after onboarding completes. Safe to call after
// construction; nil disables the post-onboarding recompute.
func (s *serviceImpl) SetRecommendationRecomputer(r RecommendationRecomputer) {
	s.recomputer = r
}

// RecomputerSetter lets the route wiring inject a recomputer into a Service
// without exposing the concrete struct.
type RecomputerSetter interface {
	SetRecommendationRecomputer(RecommendationRecomputer)
}
