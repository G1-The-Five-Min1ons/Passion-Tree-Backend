package service

import (
	"context"
	"database/sql"
	
	nodeRepo "passiontree/internal/learning-path/repository" 
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/resume/model"
	"passiontree/internal/resume/repository"
)

type serviceImpl struct {
	resumeRepo repository.ResumeRepository
	nodeRepo   nodeRepo.RepositoryNode
}

func NewService(resumeRepo repository.ResumeRepository, nodeRepo nodeRepo.RepositoryNode) ResumeService {
	return &serviceImpl{
		resumeRepo: resumeRepo,
		nodeRepo:   nodeRepo,
	}
}

type ResumeService interface {
	GetResumeNode(ctx context.Context, userID string, pathID string) (*model.ResumeResponse, error)
}

func (s *serviceImpl) GetResumeNode(ctx context.Context, userID string, pathID string) (*model.ResumeResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}
	if pathID == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}

	nodeID, err := s.resumeRepo.GetNextNodeID(ctx, userID, pathID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("No pending node found or path completed")
		}
		return nil, apperror.NewInternal(err)
	}

	nodeDetail, err := s.nodeRepo.GetNodeByID(ctx, nodeID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	return &model.ResumeResponse{
		CurrentNode: nodeDetail,
		Message:     "Resuming learning path",
	}, nil
}