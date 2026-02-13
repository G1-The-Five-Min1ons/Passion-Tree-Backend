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

	id, err := s.nodeRepo.CreateNodeWithContent(ctx, req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("node with this ID already exists")
		}

		if apperror.IsForeignKeyError(err) {
			return "", apperror.NewBadRequest("invalid path_id: learning path does not exist")
		}
		s.logger.ErrorContext(ctx, "database error during node creation", "error", err, "path_id", req.PathID)
		return "", apperror.NewInternal("failed to create node in path %s: %w", req.PathID, err)
	}

	s.logger.InfoContext(ctx, "node added successfully", "node_id", id, "path_id", req.PathID)
	return id, nil
}

func (s *serviceImpl) EditNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
	s.logger.InfoContext(ctx, "editing node details", "node_id", nodeID)

	if nodeID == "" {
		return apperror.NewBadRequest("node_id is required")
	}
	if req.Title == "" && req.Description == "" {
		return apperror.NewBadRequest("at least one field (title or description) is required for update")
	}

	if err := s.nodeRepo.UpdateNode(ctx, nodeID, req); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: node not found", "node_id", nodeID)
			return apperror.NewNotFound("cannot update: node id '%s' not found", nodeID)
		}
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("node with this title already exists in this path")
		}
		
		s.logger.ErrorContext(ctx, "database error during node update", "error", err, "node_id", nodeID)
		return apperror.NewInternal("failed to update node %s: %w", nodeID, err)
	}

	s.logger.InfoContext(ctx, "node updated successfully", "node_id", nodeID)
	return nil
}

func (s *serviceImpl) RemoveNode(ctx context.Context, nodeID string) error {

	if nodeID == "" {
		return apperror.NewBadRequest("node_id is required")
	}

	if err := s.nodeRepo.DeleteNode(ctx, nodeID); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "deletion failed: node not found", "node_id", nodeID)
			return apperror.NewNotFound("cannot delete: node id '%s' not found", nodeID)
		}
		if apperror.IsForeignKeyError(err) {
			s.logger.WarnContext(ctx, "deletion blocked: node has dependencies", "node_id", nodeID)
			return apperror.NewConflict("cannot delete node: there are existing materials, comments, or questions associated with this node")
		}
		
		s.logger.ErrorContext(ctx, "database error during node deletion", "error", err, "node_id", nodeID)
		return apperror.NewInternal("failed to remove node %s: %w", nodeID, err)
	}

	s.logger.InfoContext(ctx, "node removed successfully", "node_id", nodeID)
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
			s.logger.WarnContext(ctx, "foreign key violation: node not found", "node_id", req.NodeID)
			return "", apperror.NewBadRequest("invalid node_id: node does not exist")
		}
		
		s.logger.ErrorContext(ctx, "database error during material creation", "error", err, "node_id", req.NodeID)
		return "", apperror.NewInternal("failed to add material to node %s: %w", req.NodeID, err)
	}

	s.logger.InfoContext(ctx, "material added successfully", "material_id", id, "node_id", req.NodeID)
	return id, nil
}

func (s *serviceImpl) RemoveMaterial(ctx context.Context, materialID string) error {

	if materialID == "" {
		return apperror.NewBadRequest("material_id is required")
	}

	if err := s.nodeRepo.DeleteMaterial(ctx, materialID); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "material not found for deletion", "material_id", materialID)
			return apperror.NewNotFound("cannot delete: material id '%s' not found", materialID)
		}
		
		s.logger.ErrorContext(ctx, "database error during material deletion", "error", err, "material_id", materialID)
		return apperror.NewInternal("failed to remove material %s: %w", materialID, err)
	}

	s.logger.InfoContext(ctx, "material removed successfully", "material_id", materialID)
	return nil
}

func (s *serviceImpl) ReorderNodes(ctx context.Context, pathID string, req model.ReorderNodesRequest) error {

	if len(req.NodeIDs) == 0 {
		return apperror.NewBadRequest("node_ids list cannot be empty")
	}

	if err := s.nodeRepo.UpdateNodeSequence(ctx, req.NodeIDs); err != nil {
		s.logger.ErrorContext(ctx, "failed to update node sequence", "error", err, "path_id", pathID)
		return apperror.NewInternal("failed to reorder nodes for path %s: %w", pathID, err)
	}

	s.logger.InfoContext(ctx, "nodes sequence updated successfully", "path_id", pathID)
	return nil
}

func (s *serviceImpl) GetNodeDetails(ctx context.Context, nodeID string) (*model.Node, error) {

	if nodeID == "" {
		return nil, apperror.NewBadRequest("node_id is required")
	}
	
	node, err := s.nodeRepo.GetNodeByID(ctx, nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "node details not found", "node_id", nodeID)
			return nil, apperror.NewNotFound("node with id '%s' not found", nodeID)
		}
		
		s.logger.ErrorContext(ctx, "database error fetching node details", "error", err, "node_id", nodeID)
		return nil, apperror.NewInternal("failed to retrieve details for node %s: %w", nodeID, err)
	}

	s.logger.InfoContext(ctx, "node details retrieved successfully", "node_id", nodeID)
	return node, nil
}

func (s *serviceImpl) GetNodesByPathID(ctx context.Context, pathID string) ([]model.Node, error) {

	if pathID == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}

	nodes, err := s.nodeRepo.GetNodesByPathID(ctx, pathID)
	if err != nil {
		s.logger.ErrorContext(ctx, "database error fetching nodes by path id", "error", err, "path_id", pathID)
		return nil, apperror.NewInternal("failed to retrieve nodes for path %s: %w", pathID, err)
	}

	if nodes == nil {
		nodes = []model.Node{}
	}

	s.logger.InfoContext(ctx, "nodes retrieved successfully for path", "path_id", pathID, "count", len(nodes))
	return nodes, nil
}
