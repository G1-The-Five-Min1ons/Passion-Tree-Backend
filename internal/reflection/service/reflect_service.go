package service

import (
	"context"
	"database/sql"
	"fmt"
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
		return nil, apperror.NewInternal(fmt.Errorf("AI analysis service is not configured"))
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
		return nil, apperror.NewInternal(fmt.Errorf("failed to analyze reflection: %w", err))
	}

	s.logger.InfoContext(ctx, "AI analysis successful", 
		"sentiment", sentimentResp.SentimentAnalysis, 
		"reflection_score", sentimentResp.ReflectionScore,
		"weighted_reflection_score", sentimentResp.WeightedReflectionScore,
	)

	id, err := s.refRepo.CreateReflection(ctx, req, sentimentResp.Summary, sentimentResp.SentimentAnalysis, sentimentResp.PrimaryEmotion, sentimentResp.StrugglePoint, sentimentResp.AIConfidentScore, sentimentResp.ReflectionScore, sentimentResp.WeightedReflectionScore)
	if err != nil {
		fmt.Printf("CreateReflection database error: %v\n", err)
		if apperror.IsDuplicateKeyError(err) {
			return nil, apperror.NewConflict("reflection with this ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid tree_node_id or user_id: node or user does not exist")
		}
		return nil, apperror.NewInternal(err)
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
	if reflectID == "" {
		return nil, apperror.NewBadRequest("reflect_id is required")
	}
	ref, err := s.refRepo.GetReflectionByID(ctx, reflectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("reflection with id '%s' not found", reflectID)
		}
		return nil, apperror.NewInternal(err)
	}
	return ref, nil
}

func (s *serviceImpl) GetAllReflections(ctx context.Context) ([]model.Reflection, error) {
	reflections, err := s.refRepo.GetAllReflections(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return reflections, nil
}

func (s *serviceImpl) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
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
			return apperror.NewNotFound("cannot update: reflection id '%s' not found", reflectID)
		}
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("reflection with this information already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return apperror.NewBadRequest("invalid tree_node_id: node does not exist")
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) DeleteReflection(ctx context.Context, reflectID string) error {
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
		return apperror.NewInternal(err)
	}

	s.logger.InfoContext(ctx, "reflection deleted successfully", "reflect_id", reflectID)
	return nil
}
