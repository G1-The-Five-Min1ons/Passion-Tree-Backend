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
		sentimentReq := &aiclient.SentimentRequest{
			WhatLearned:           req.Learned,
			FeelingsAfterLearning: req.Reflect,
		}

		sentimentResp, err := s.aiClient.AnalyzeSentiment(ctx, *sentimentReq)
		if err != nil {
			fmt.Printf("AI sentiment analysis failed: %v\n", err)
		} else {
			req.Mood = sentimentResp.Sentiment
			if req.Tag == "" {
				req.Tag = sentimentResp.Advanced.PrimaryEmotion
			}
			fmt.Printf("AI Sentiment: %s, Score: %.2f, Summary: %s\n",
				sentimentResp.Sentiment, sentimentResp.ReflectionScore, sentimentResp.Summary)
		}
	}

	id, err := s.refRepo.CreateReflection(ctx, req)
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
		ID:        id,
		Score:     req.FeelScore,
		Mood:      req.Mood,
		Summary:   req.Learned,
		CreatedAt: "",
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
	return nil
}
