package service

import (
	"context"
	"database/sql"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

// CreateTreeNode creates a new tree node
func (s *serviceImpl) CreateTreeNode(ctx context.Context, req model.CreateTreeNodeRequest) (*model.TreeNodeResponse, error) {
	s.logger.InfoContext(ctx, "creating new tree node", "tree_id", req.TreeID, "node_id", req.NodeID)

	// Validate required fields
	if req.NodeTitle == "" {
		return nil, apperror.NewBadRequest("node_title is required")
	}
	if req.NodeID == "" {
		return nil, apperror.NewBadRequest("node_id is required")
	}
	if req.TreeID == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}
	
	// Create tree node
	treeNodeID, err := s.refRepo.CreateTreeNode(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create tree node", "error", err)
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid tree_id or node_id: referenced resource does not exist")
		}
		return nil, apperror.NewInternal(err)
	}
	
	// Get the created tree node
	treeNode, err := s.refRepo.GetTreeNodeByID(ctx, treeNodeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tree node after creation", "error", err, "tree_node_id", treeNodeID)
		return nil, apperror.NewInternal(err)
	}

	s.logger.InfoContext(ctx, "tree node created successfully", "tree_node_id", treeNodeID)
	
	return &model.TreeNodeResponse{
		TreeNodeID: treeNode.TreeNodeID,
		NodeTitle:  treeNode.NodeTitle,
		NodeID:     treeNode.NodeID,
		NodeScore:  treeNode.NodeScore,
		CreatedAt:  treeNode.CreatedAt,
		TreeID:     treeNode.TreeID,
		ChildNode:  treeNode.ChildNode,
		Sequence:   treeNode.Sequence,
	}, nil
}

// GetTreeNodesByTreeID retrieves all tree nodes for a specific tree
func (s *serviceImpl) GetTreeNodesByTreeID(ctx context.Context, treeID string) ([]model.TreeNode, error) {
	s.logger.InfoContext(ctx, "fetching tree nodes", "tree_id", treeID)

	if treeID == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}
	
	nodes, err := s.refRepo.GetTreeNodesByTreeID(ctx, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tree nodes", "error", err, "tree_id", treeID)
		return nil, apperror.NewInternal(err)
	}
	
	if len(nodes) == 0 {
		s.logger.InfoContext(ctx, "no tree nodes found", "tree_id", treeID)
		return []model.TreeNode{}, nil
	}
	
	s.logger.InfoContext(ctx, "successfully retrieved tree nodes", "tree_id", treeID, "count", len(nodes))
	return nodes, nil
}

// GetTreeNodeByID retrieves a specific tree node by ID
func (s *serviceImpl) GetTreeNodeByID(ctx context.Context, treeNodeID string) (*model.TreeNode, error) {
	s.logger.InfoContext(ctx, "fetching tree node by ID", "tree_node_id", treeNodeID)

	if treeNodeID == "" {
		return nil, apperror.NewBadRequest("tree_node_id is required")
	}
	
	node, err := s.refRepo.GetTreeNodeByID(ctx, treeNodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "tree node not found", "tree_node_id", treeNodeID)
			return nil, apperror.NewNotFound("tree node with id '%s' not found", treeNodeID)
		}
		s.logger.ErrorContext(ctx, "database error fetching tree node", "error", err, "tree_node_id", treeNodeID)
		return nil, apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "successfully retrieved tree node", "tree_node_id", treeNodeID)
	return node, nil
}

// UpdateTreeNode updates an existing tree node
func (s *serviceImpl) UpdateTreeNode(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error {
	s.logger.InfoContext(ctx, "updating tree node", "tree_node_id", treeNodeID)

	if treeNodeID == "" {
		return apperror.NewBadRequest("tree_node_id is required")
	}
	if req.NodeTitle == "" {
		return apperror.NewBadRequest("node_title is required")
	}
	
	err := s.refRepo.UpdateTreeNode(ctx, treeNodeID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: tree node not found", "tree_node_id", treeNodeID)
			return apperror.NewNotFound("tree node with id '%s' not found", treeNodeID)
		}
		s.logger.ErrorContext(ctx, "failed to update tree node", "error", err, "tree_node_id", treeNodeID)
		return apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "tree node updated successfully", "tree_node_id", treeNodeID)
	return nil
}

// DeleteTreeNode deletes a tree node
func (s *serviceImpl) DeleteTreeNode(ctx context.Context, treeNodeID string) error {
	s.logger.InfoContext(ctx, "request to delete tree node", "tree_node_id", treeNodeID)

	if treeNodeID == "" {
		return apperror.NewBadRequest("tree_node_id is required")
	}
	
	err := s.refRepo.DeleteTreeNode(ctx, treeNodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("tree node with id '%s' not found", treeNodeID)
		}
		s.logger.ErrorContext(ctx, "failed to delete tree node", "error", err, "tree_node_id", treeNodeID)
		return apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "tree node deleted successfully", "tree_node_id", treeNodeID)
	return nil
}
