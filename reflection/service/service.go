package service

import (
	"context"
	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
)

type ServiceReflection interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error)
	GetReflection(ctx context.Context, id string) (*model.Reflection, error)
	ListReflections(ctx context.Context, nodeID string) ([]model.Reflection, error)
	DeleteReflection(ctx context.Context, id string) error
}

type Service interface {
	ServiceReflection
}

type serviceImpl struct {
	refRepo repository.RepositoryReflection
}

func NewService(repo repository.RepositoryReflection) Service {
	return &serviceImpl{
		refRepo: repo,
	}
}

func (s *serviceImpl) CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error) {

//Validation
	if strings.TrimSpace(req.Learned) == "" {
		return nil, errors.New("What have learned is required")
	}

	if strings.TrimSpace(req.Reflect) == "" {
		return nil, errors.New("Reflection is required")
	}

	if req.FeelScore == 0 {
		return nil, errors.New("Emotion is required")
	}

	if req.ProgressScore == 0 {
		return nil, errors.New("Progress score is required")
	}

	if req.ChallengeScore == 0 {
		return nil, errors.New("Challenge score is required")
	}

	if strings.TrimSpace(req.TreeNodeID) == "" {
		return nil, errors.New("tree node_id is required")
	}

	ref := model.Reflection{
    ReflectID:          uuid.New(),
    ReflectScore:       req.FeelScore,
    ReflectDescription: req.Learned,
    Reflect:            req.Reflect,
    Mood:               "",
    Tag:                "",
    ProgressScore:      req.ProgressScore,
    ChallengeScore:     req.ChallengeScore,
    CreatedAt:          time.Now(),
    TreeNodeID:         uuid.MustParse(req.TreeNodeID),
	}

	if err := s.refRepo.Save(ctx, ref); err != nil {
		return nil, err
	}

	return &model.ReflectionResponse{
		ID: ref.ReflectID.String(),
		// …
	}, nil
}

func (s *serviceImpl) UpdateReflection(ctx context.Context, id string, req model.UpdateReflectionRequest) error {

    refID, err := uuid.Parse(id)
    if err != nil {
        return errors.New("invalid reflect id")
    }

    if strings.TrimSpace(req.Learned) == "" {
        return errors.New("What have learned required")
    }

    if strings.TrimSpace(req.Reflect) == "" {
        return errors.New("Reflection required")
    }

    if req.FeelScore == 0{
        return errors.New("Emotion required")
    }

	if req.ProgressScore ==0 {
		return errors.New("Progress score required")
	}

	if req.ChallengeScore ==0 {
		return errors.New("Challenge score required")
	}

    update := model.Reflection{
        ReflectID:        refID,
        ReflectScore:     req.FeelScore,
        ReflectDescription: req.Learned,
        Reflect:          req.Reflect,
        Mood:             req.Mood,
        Tag:              req.Tag,
        ProgressScore:    req.ProgressScore,
        ChallengeScore:   req.ChallengeScore,
    }

    return s.refRepo.Update(ctx, update)
}

func (s *serviceImpl) DeleteReflection(ctx context.Context, id string) error {

    refID, err := uuid.Parse(id)
    if err != nil {
        return err
    }

    ref := model.Reflection{
        ReflectID: refID,
    }

    return s.refRepo.Delete(ctx, ref)
}
