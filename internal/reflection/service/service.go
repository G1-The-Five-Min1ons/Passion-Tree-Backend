package service

import (
	"context"
	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
	"passiontree/internal/pkg/reflecterror"
)

type ReflectionService interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error)
	GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error)
	GetAllReflections(ctx context.Context) ([]model.Reflection, error)
	UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflection(ctx context.Context, reflectID string) error
}

type serviceImpl struct {
	refRepo repository.RepositoryReflection
}

func NewService(repo repository.RepositoryReflection) ReflectionService {
	return &serviceImpl{
		refRepo: repo,
	}
}

func (s *serviceImpl) CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error) {
	//Validation
	if strings.TrimSpace(req.Learned) == "" {
		return nil, apperor.NewBadRequest("what have learned is required")
	}

	if strings.TrimSpace(req.Reflect) == "" {
		return nil, apperor.NewBadRequest("reflection is required")
	}

	if strings.TrimSpace(req.FeelScore) == "" {
		return nil, apperor.NewBadRequest("feel_score is required")
	}

	if strings.TrimSpace(req.ProgressScore) == "" {
		return nil, apperor.NewBadRequest("progress_score is required")
	}

	if strings.TrimSpace(req.ChallengeScore) == "" {
		return nil, apperor.NewBadRequest("challenge_score is required")
	}

	if strings.TrimSpace(req.TreeNodeID) == "" {
		return nil, apperor.NewBadRequest("tree_node_id is required")
	}

	id, err := s.refRepo.CreateReflection(ctx, req)
	if err != nil {
		return nil, err
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
		return nil, apperor.NewBadRequest("reflect_id is required")
	}

	ref, err := s.refRepo.GetReflectionByID(ctx, reflectID)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (s *serviceImpl) GetAllReflections(ctx context.Context) ([]model.Reflection, error) {
	reflections, err := s.refRepo.GetAllReflections(ctx)
	if err != nil {
		return nil, err
	}

	return reflections, nil
}

func (s *serviceImpl) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
	if strings.TrimSpace(reflectID) == "" {
		return apperor.NewBadRequest("reflect_id is required")
	}

	if strings.TrimSpace(req.Learned) == "" {
		return apperor.NewBadRequest("what have learned is required")
	}

	if strings.TrimSpace(req.Reflect) == "" {
		return apperor.NewBadRequest("reflection is required")
	}

	if strings.TrimSpace(req.FeelScore) == "" {
		return apperor.NewBadRequest("feel_score is required")
	}

	if strings.TrimSpace(req.ProgressScore) == "" {
		return apperor.NewBadRequest("progress_score is required")
	}

	if strings.TrimSpace(req.ChallengeScore) == "" {
		return apperor.NewBadRequest("challenge_score is required")
	}

	if err := s.refRepo.UpdateReflection(ctx, reflectID, req); err != nil {
		return err
	}

	return nil
}

func (s *serviceImpl) DeleteReflection(ctx context.Context, reflectID string) error {
	if strings.TrimSpace(reflectID) == "" {
		return apperor.NewBadRequest("reflect_id is required")
	}

	if err := s.refRepo.DeleteReflection(ctx, reflectID); err != nil {
		return err
	}

	return nil
}
