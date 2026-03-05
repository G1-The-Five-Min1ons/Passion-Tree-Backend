package service

import (
	"context"
	"database/sql"
	
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

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
		return nil, apperror.NewInternal("failed to get resume node: %w", err)
	}

	nodeDetail, err := s.nodeRepo.GetNodeByID(ctx, nodeID, userID)
	if err != nil {
		return nil, apperror.NewInternal("failed to get node detail: %w", err)
	}

	s.logger.InfoContext(ctx, "resume node fetched successfully", "user_id", userID, "path_id", pathID, "node_id", nodeID)
	return &model.ResumeResponse{
		CurrentNode: nodeDetail,
		Message:     "Resuming learning path",
	}, nil
}