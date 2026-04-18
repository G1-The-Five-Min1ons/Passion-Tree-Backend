package service

import (
	"context"
	recmodel "passiontree/internal/recommendation/model"
)

func (s *serviceImpl) RunDailyRecommendationBatch(ctx context.Context) error {
	s.logger.InfoContext(ctx, "starting daily recommendation batch...")

	interactions, err := s.recRepo.GetBatchInteractions(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get batch interactions", "error", err)
		return err
	}

	profiles, err := s.recRepo.GetBatchProfiles(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get batch profiles", "error", err)
		return err
	}

	payload := recmodel.BatchRecommendPayload{
		Interactions: interactions,
		Profiles:     profiles,
	}

	s.logger.InfoContext(ctx, "sending data to AI engine", "interactions_count", len(interactions), "profiles_count", len(profiles))

	aiResp, err := s.aiClient.ComputeBatchRecommendations(ctx, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "AI engine computation failed", "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "batch calculation completed successfully!", "processed_users", len(aiResp.Data))

	s.logger.InfoContext(ctx, "saving AI recommendations to database...", "user_count", len(aiResp.Data))

	err = s.recRepo.SaveBatchRecommendations(ctx, aiResp.Data)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to save recommendations to database", "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "daily recommendation batch completed successfully")

	return nil
}
