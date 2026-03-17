package service

import (
	"context"
	"database/sql"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/reflection/model"
)

func (s *serviceImpl) CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error) {

	if req.LearningReflect == "" {
		return nil, apperror.NewBadRequest("learning_reflect is required")
	}
	if req.MoodReflect == "" {
		return nil, apperror.NewBadRequest("mood_reflect is required")
	}
	if req.TreeNodeID == "" {
		return nil, apperror.NewBadRequest("tree_node_id is required")
	}

	// AI analysis is required
	if s.aiClient == nil {
		s.logger.ErrorContext(ctx, "AI client is not available")
		return nil, apperror.NewInternal("AI analysis service is not configured")
	}

	s.logger.InfoContext(ctx, "calling AI sentiment analysis service")

	sentimentReq := &aiclient.SentimentRequest{
		LearningReflect: req.LearningReflect,
		MoodReflect:     req.MoodReflect,
		FeelScore:       req.FeelScore,
		ProgressScore:   req.ProgressScore,
		ChallengeScore:  req.ChallengeScore,
	}

	sentimentResp, err := s.aiClient.AnalyzeSentiment(ctx, *sentimentReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "AI sentiment analysis failed", "error", err)
		return nil, apperror.NewInternal("failed to analyze reflection: %v", err)
	}

	s.logger.InfoContext(ctx, "AI analysis successful",
		"sentiment", sentimentResp.SentimentAnalysis,
		"reflection_score", sentimentResp.ReflectionScore,
		"weighted_reflection_score", sentimentResp.WeightedReflectionScore,
	)

	id, err := s.refRepo.CreateReflection(ctx, req, sentimentResp.Summary, sentimentResp.SentimentAnalysis, sentimentResp.PrimaryEmotion, sentimentResp.StrugglePoint, sentimentResp.AIConfidentScore, sentimentResp.ReflectionScore, sentimentResp.WeightedReflectionScore)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to save reflection to database",
			"error", err,
			"tree_node_id", req.TreeNodeID,
		)
		if apperror.IsDuplicateKeyError(err) {
			return nil, apperror.NewConflict("reflection with this ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid tree_node_id or user_id: node or user does not exist")
		}
		return nil, apperror.NewInternal("failed to create reflection: %w", err)
	}

	s.logger.InfoContext(ctx, "reflection created successfully", "reflection_id", id)

	// Recalculate tree score (best-effort: a failure here must not reject the reflection).
	if node, getErr := s.refRepo.GetTreeNodeByID(ctx, req.TreeNodeID); getErr == nil && node != nil {
		if _, scoreErr := s.refRepo.CalculateAndUpdateTreeScore(ctx, node.TreeID); scoreErr != nil {
			s.logger.WarnContext(ctx, "failed to recalculate tree score after reflection",
				"tree_id", node.TreeID, "error", scoreErr)
		}
	}

	return &model.ReflectionResponse{
		ReflectID:               id,
		Summary:                 sentimentResp.Summary,
		SentimentAnalysis:       sentimentResp.SentimentAnalysis,
		PrimaryEmotion:          sentimentResp.PrimaryEmotion,
		StrugglePoint:           sentimentResp.StrugglePoint,
		DevelopmentPlan:         sentimentResp.DevelopmentPlan,
		AIConfidentScore:        sentimentResp.AIConfidentScore,
		ReflectionScore:         sentimentResp.ReflectionScore,
		WeightedReflectionScore: sentimentResp.WeightedReflectionScore,
	}, nil
}

func (s *serviceImpl) GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error) {
	s.logger.InfoContext(ctx, "fetching reflection by ID", "reflect_id", reflectID)

	if reflectID == "" {
		return nil, apperror.NewBadRequest("reflect_id is required")
	}
	ref, err := s.refRepo.GetReflectionByID(ctx, reflectID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "reflection not found", "reflect_id", reflectID)
			return nil, apperror.NewNotFound("reflection with id '%s' not found", reflectID)
		}
		s.logger.ErrorContext(ctx, "database error fetching reflection", "error", err, "reflect_id", reflectID)
		return nil, apperror.NewInternal("database error fetching reflection: %w", err)
	}

	s.logger.InfoContext(ctx, "successfully retrieved reflection", "reflect_id", reflectID)

	return ref, nil
}

func (s *serviceImpl) GetAllReflections(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
	s.logger.InfoContext(ctx, "fetching reflections with filters",
		"tree_node_id", filter.TreeNodeID,
		"tree_id", filter.TreeID,
		"album_id", filter.AlbumID,
		"user_id", filter.UserID,
		"before_created_at", filter.BeforeCreatedAt,
		"before_reflect_id", filter.BeforeReflectID,
		"limit", filter.Limit,
	)

	reflections, err := s.refRepo.GetAllReflections(ctx, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch reflections", "error", err)
		return nil, apperror.NewInternal("failed to fetch reflections: %w", err)
	}
	s.logger.InfoContext(ctx, "successfully fetched reflections", "count", len(reflections))
	return reflections, nil
}

func (s *serviceImpl) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
	s.logger.InfoContext(ctx, "updating reflection", "reflect_id", reflectID)

	if reflectID == "" {
		return apperror.NewBadRequest("reflect_id is required")
	}
	if req.LearningReflect == "" {
		return apperror.NewBadRequest("learning_reflect is required")
	}
	if req.MoodReflect == "" {
		return apperror.NewBadRequest("mood_reflect is required")
	}

	if err := s.refRepo.UpdateReflection(ctx, reflectID, req); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: reflection not found", "reflect_id", reflectID)
			return apperror.NewNotFound("cannot update: reflection id '%s' not found", reflectID)
		}
		if apperror.IsDuplicateKeyError(err) {
			s.logger.ErrorContext(ctx, "failed to update reflection: duplicate key", "reflect_id", reflectID, "error", err)
			return apperror.NewConflict("reflection with this information already exists")
		}
		if apperror.IsForeignKeyError(err) {
			s.logger.ErrorContext(ctx, "failed to update reflection: foreign key error", "reflect_id", reflectID, "error", err)
			return apperror.NewBadRequest("invalid tree_node_id: node does not exist")
		}
		return apperror.NewInternal("failed to update reflection: %w", err)
	}
	s.logger.InfoContext(ctx, "reflection updated successfully", "reflect_id", reflectID)
	return nil
}

func (s *serviceImpl) DeleteReflection(ctx context.Context, reflectID string) error {
	s.logger.InfoContext(ctx, "request to delete reflection", "reflect_id", reflectID)

	if reflectID == "" {
		return apperror.NewBadRequest("reflect_id is required")
	}
	if err := s.refRepo.DeleteReflection(ctx, reflectID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("reflection with id '%s' not found", reflectID)
		}
		if apperror.IsForeignKeyError(err) {
			return apperror.NewConflict("cannot delete reflection: there are existing dependencies associated with this reflection")
		}
		s.logger.ErrorContext(ctx, "failed to delete reflection", "error", err, "reflect_id", reflectID)
		return apperror.NewInternal("failed to delete reflection: %w", err)
	}

	s.logger.InfoContext(ctx, "reflection deleted successfully", "reflect_id", reflectID)
	return nil
}
