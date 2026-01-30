package service

import (
	"context"
	"database/sql"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/reflection/model"
)

func (s *serviceImpl) CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error) {
	s.logger.InfoContext(ctx, "starting reflection creation", "tree_node_id", req.TreeNodeID)
	
	if req.Learned == "" {
		return nil, apperror.NewBadRequest("what have learned is required")
	}
	if req.Reflect == "" {
		return nil, apperror.NewBadRequest("reflection is required")
	}
	if req.FeelScore == "" {
		return nil, apperror.NewBadRequest("feel_score is required")
	}
	if req.ProgressScore == "" {
		return nil, apperror.NewBadRequest("progress_score is required")
	}
	if req.ChallengeScore == "" {
		return nil, apperror.NewBadRequest("challenge_score is required")
	}
	if req.TreeNodeID == "" {
		return nil, apperror.NewBadRequest("tree_node_id is required")
	}

	if s.aiClient != nil {
		s.logger.InfoContext(ctx, "calling AI sentiment analysis service")

		sentimentReq := &aiclient.SentimentRequest{
			WhatLearned:           req.Learned,
			FeelingsAfterLearning: req.Reflect,
		}

		sentimentResp, err := s.aiClient.AnalyzeSentiment(ctx, *sentimentReq)
		if err != nil {
			s.logger.WarnContext(ctx, "AI sentiment analysis failed, proceeding with defaults", "error", err)
		} else {
			req.Mood = sentimentResp.Sentiment
			if req.Tag == "" {
				req.Tag = sentimentResp.Advanced.PrimaryEmotion
			}
			s.logger.InfoContext(ctx, "AI analysis successful", 
				"sentiment", sentimentResp.Sentiment, 
				"reflection_score", sentimentResp.ReflectionScore,
			)
		}
	}

	id, err := s.refRepo.CreateReflection(ctx, req)
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
		return nil, apperror.NewInternal(err)
	}

	s.logger.InfoContext(ctx, "reflection created successfully", "reflection_id", id)

	return &model.ReflectionResponse{
		ID:        id,
		Score:     req.FeelScore,
		Mood:      req.Mood,
		Summary:   req.Learned,
		CreatedAt: "",
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
		return nil, apperror.NewInternal(err)
	}

	s.logger.InfoContext(ctx, "successfully retrieved reflection", "reflect_id", reflectID)

	return ref, nil
}

func (s *serviceImpl) GetAllReflections(ctx context.Context) ([]model.Reflection, error) {
	reflections, err := s.refRepo.GetAllReflections(ctx)
	s.logger.InfoContext(ctx, "fetching all reflections", "count", len(reflections))

	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch all reflections", "error", err)
		return nil, apperror.NewInternal(err)
	}
	s.logger.InfoContext(ctx, "successfully fetched reflections", "count", len(reflections))
	return reflections, nil
 }

func (s *serviceImpl) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
	s.logger.InfoContext(ctx, "updating reflection", "reflect_id", reflectID)
	
	if reflectID == "" {
		return apperror.NewBadRequest("reflect_id is required")
	}
	if req.Learned == "" {
		return apperror.NewBadRequest("what have learned is required")
	}
	if req.Reflect == "" {
		return apperror.NewBadRequest("reflection is required")
	}
	if req.FeelScore == "" {
		return apperror.NewBadRequest("feel_score is required")
	}
	if req.ProgressScore == "" {
		return apperror.NewBadRequest("progress_score is required")
	}
	if req.ChallengeScore == "" {
		return apperror.NewBadRequest("challenge_score is required")
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
		return apperror.NewInternal(err)
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
		return apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "reflection deleted successfully", "reflect_id", reflectID)
	return nil
}
