package service

import (
	"context"
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) AddNode(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	if req.Title == "" {
		return "", apperror.NewBadRequest("node title is required")
	}

	id, err := s.nodeRepo.CreateNode(ctx, req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("node with this ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return "", apperror.NewBadRequest("invalid path_id: learning path does not exist")
		}
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) EditNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
	if nodeID == "" {
		return apperror.NewBadRequest("node_id is required")
	}
	if req.Title == "" &&
		req.Description == "" {
		return apperror.NewBadRequest("request is required")
	}
	if err := s.nodeRepo.UpdateNode(ctx, nodeID, req); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot update: node id '%s' not found", nodeID)
		}
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("node with this title already exists in this path")
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) RemoveNode(ctx context.Context, nodeID string) error {
	if nodeID == "" {
		return apperror.NewBadRequest("node_id is required")
	}
	if err := s.nodeRepo.DeleteNode(ctx, nodeID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot delete: node id '%s' not found", nodeID)
		}
		if apperror.IsForeignKeyError(err) {
			return apperror.NewConflict("cannot delete node: there are existing materials, comments, or questions associated with this node")
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) AddMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error) {
	if req.Type == "" || req.URL == "" {
		return "", apperror.NewBadRequest("material type and url are required")
	}

	id, err := s.nodeRepo.CreateMaterial(ctx, req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("material with this ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return "", apperror.NewBadRequest("invalid node_id: node does not exist")
		}
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) RemoveMaterial(ctx context.Context, materialID string) error {
	if materialID == "" {
		return apperror.NewBadRequest("material_id is required")
	}
	if err := s.nodeRepo.DeleteMaterial(ctx, materialID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot delete: material id '%s' not found", materialID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) ReorderNodes(ctx context.Context, pathID string, req model.ReorderNodesRequest) error {
	for index, nodeID := range req.NodeIDs {
		err := s.nodeRepo.UpdateNodeSequence(ctx, nodeID, index)
		if err != nil {
			return apperror.NewInternal(err)
		}
	}
	return nil
}


func (s *serviceImpl) GetNodeDetails(ctx context.Context, nodeID string) (*model.Node, error) {
	if nodeID == "" {
		return nil, apperror.NewBadRequest("node_id is required")
	}
	
	node, err := s.nodeRepo.GetNodeByID(ctx, nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("node with id '%s' not found", nodeID)
		}
		return nil, apperror.NewInternal(err)
	}

	return node, nil
}