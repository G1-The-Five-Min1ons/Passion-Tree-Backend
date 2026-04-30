package service

import (
	"context"
	"errors"
	recmodel "passiontree/internal/recommendation/model"
)

// RecomputeForUser computes and saves personalized recommendations for a single
// user without waiting for the daily batch. Designed to be called right after
// the user finishes onboarding so they get a personalized home feed instead of
// the popular-paths fallback. Returns nil (no-op) if the user has no profile.
func (s *serviceImpl) RecomputeForUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user_id is required")
	}
	if s.aiClient == nil {
		s.logger.ErrorContext(ctx, "recompute aborted: AI client is nil", "user_id", userID)
		return errors.New("ai client is not initialized")
	}

	profile, err := s.recRepo.GetUserProfile(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user profile for recompute", "user_id", userID, "error", err)
		return err
	}
	if profile == nil {
		s.logger.InfoContext(ctx, "skipping recompute: user has no onboarding profile yet", "user_id", userID)
		return nil
	}

	interactions, err := s.recRepo.GetUserInteractions(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user interactions for recompute", "user_id", userID, "error", err)
		return err
	}

	payload := recmodel.BatchRecommendPayload{
		Interactions: interactions,
		Profiles:     []recmodel.UserProfile{*profile},
	}

	s.logger.InfoContext(ctx, "recomputing recommendations for single user",
		"user_id", userID,
		"interactions_count", len(interactions),
	)

	aiResp, err := s.aiClient.ComputeBatchRecommendations(ctx, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "AI recompute failed", "user_id", userID, "error", err)
		return err
	}

	if len(aiResp.Data) == 0 {
		s.logger.WarnContext(ctx, "AI returned no recommendations for user", "user_id", userID)
		return nil
	}

	if err := s.recRepo.SaveBatchRecommendations(ctx, aiResp.Data); err != nil {
		s.logger.ErrorContext(ctx, "failed to save recompute result", "user_id", userID, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "single-user recommendation recompute completed", "user_id", userID)
	return nil
}

func (s *serviceImpl) RunDailyRecommendationBatch(ctx context.Context) error {
	s.logger.InfoContext(ctx, "starting daily recommendation batch...")

	if s.aiClient == nil {
		s.logger.ErrorContext(ctx, "batch aborted: AI client is nil")
		return errors.New("ai client is not initialized")
	}

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
