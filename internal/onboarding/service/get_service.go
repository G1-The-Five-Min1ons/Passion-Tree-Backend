package service

import (
	"context"

	"passiontree/internal/onboarding/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) GetOnboarding(ctx context.Context, userID string) (*model.OnboardingData, error) {
	data, err := s.repo.GetOnboardingByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("failed to get onboarding: %v", err)
	}
	return data, nil
}
