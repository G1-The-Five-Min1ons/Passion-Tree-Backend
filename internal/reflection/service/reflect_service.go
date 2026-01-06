package service

import (
	"context"
	"database/sql"
	"strings"
	"passiontree/internal/reflection/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error) {
	if strings.TrimSpace(req.Learned) == "" {
		return nil, apperror.NewBadRequest("what have learned is required")
	}
	if strings.TrimSpace(req.Reflect) == "" {
		return nil, apperror.NewBadRequest("reflection is required")
	}
	if strings.TrimSpace(req.FeelScore) == "" {
		return nil, apperror.NewBadRequest("feel_score is required")
	}
	if strings.TrimSpace(req.ProgressScore) == "" {
		return nil, apperror.NewBadRequest("progress_score is required")
	}
	if strings.TrimSpace(req.ChallengeScore) == "" {
		return nil, apperror.NewBadRequest("challenge_score is required")
	}
	if strings.TrimSpace(req.TreeNodeID) == "" {
		return nil, apperror.NewBadRequest("tree_node_id is required")
	}
	id, err := s.refRepo.CreateReflection(ctx, req)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &model.ReflectionResponse{
		ID:        id,
		Score:     req.FeelScore,
		Mood:      "",
		Summary:   req.Learned,
		CreatedAt: "",
	}, nil
}

func (s *serviceImpl) GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error) {
	if strings.TrimSpace(reflectID) == "" {
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
	if strings.TrimSpace(reflectID) == "" {
		return apperror.NewBadRequest("reflect_id is required")
	}
	if strings.TrimSpace(req.Learned) == "" {
		return apperror.NewBadRequest("what have learned is required")
	}
	if strings.TrimSpace(req.Reflect) == "" {
		return apperror.NewBadRequest("reflection is required")
	}
	if strings.TrimSpace(req.FeelScore) == "" {
		return apperror.NewBadRequest("feel_score is required")
	}
	if strings.TrimSpace(req.ProgressScore) == "" {
		return apperror.NewBadRequest("progress_score is required")
	}
	if strings.TrimSpace(req.ChallengeScore) == "" {
		return apperror.NewBadRequest("challenge_score is required")
	}
	if err := s.refRepo.UpdateReflection(ctx, reflectID, req); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot update: reflection id '%s' not found", reflectID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) DeleteReflection(ctx context.Context, reflectID string) error {
	if strings.TrimSpace(reflectID) == "" {
		return apperror.NewBadRequest("reflect_id is required")
	}
	if err := s.refRepo.DeleteReflection(ctx, reflectID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("reflection with id '%s' not found", reflectID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}
