package service

import (
	"context"
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
	missionModel "passiontree/internal/mission/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/utils"
	"time"
)

func (s *serviceImpl) AddNode(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	if req.Title == "" {
		return "", apperror.NewBadRequest("node title is required")
	}

	if req.Link_vdo != "" {
		s.logger.InfoContext(ctx, "Link_vdo = ", "Link_vdo", req.Link_vdo)
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
	if req.Title == "" && req.Description == "" && req.Link_vdo == "" && req.Materials == nil {
		return apperror.NewBadRequest("at least one field (title, description, link_vdo, or materials) is required for update")
	}

	existingNode, err := s.nodeRepo.GetNodeByID(ctx, nodeID, "")

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: node not found", "node_id", nodeID)
			return apperror.NewNotFound("cannot update: node id '%s' not found", nodeID)
		}

		s.logger.ErrorContext(ctx, "database error during node update", "error", err, "node_id", nodeID)
		return apperror.NewInternal("failed to update node %s: %w", nodeID, err)
	}

	if req.Title == "" {
		req.Title = existingNode.Title
	}
	if req.Description == "" {
		req.Description = existingNode.Description
	}
	if req.Link_vdo == "" {
		req.Link_vdo = existingNode.Link_vdo
	}

	if err := s.nodeRepo.UpdateNode(ctx, nodeID, req); err != nil {

		s.logger.ErrorContext(ctx, "failed to update node", "error", err, "node_id", nodeID)
		return apperror.NewInternal("failed to update node: %w", err)
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

func (s *serviceImpl) GetNodeDetails(ctx context.Context, nodeID string, userID string) (*model.Node, error) {

	if nodeID == "" {
		return nil, apperror.NewBadRequest("node_id is required")
	}

	node, err := s.nodeRepo.GetNodeByID(ctx, nodeID, userID)
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

func (s *serviceImpl) GetNodesByPathID(ctx context.Context, pathID string, userID string) ([]model.Node, error) {

	if pathID == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}

	nodes, err := s.nodeRepo.GetNodesByPathID(ctx, pathID, userID)
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

func (s *serviceImpl) StartNode(ctx context.Context, nodeID string, userID string) error {
	if nodeID == "" || userID == "" {
		return apperror.NewBadRequest("node_id and user_id are required")
	}

	if err := s.nodeRepo.UpdateNodeProgressStatus(ctx, nodeID, userID); err != nil {
		s.logger.ErrorContext(ctx, "failed to start node", "error", err, "node_id", nodeID, "user_id", userID)
		return apperror.NewInternal("failed to start node: %w", err)
	}

	s.logger.InfoContext(ctx, "node started successfully", "node_id", nodeID, "user_id", userID)
	return nil
}

func (s *serviceImpl) CompleteNode(ctx context.Context, nodeID string, userID string) error {
	if nodeID == "" || userID == "" {
		return apperror.NewBadRequest("node_id and user_id are required")
	}

	if err := s.nodeRepo.UpdateNodeProgressCompletion(ctx, nodeID, userID); err != nil {
		s.logger.ErrorContext(ctx, "failed to complete node", "error", err, "node_id", nodeID, "user_id", userID)
		return apperror.NewInternal("failed to complete node: %w", err)
	}

	// Award XP for completing a node
	if err := s.xpRepo.AddXPAndRecalcLevel(ctx, userID, repository.XPPerNodeComplete); err != nil {
		s.logger.ErrorContext(ctx, "failed to award XP for node completion", "error", err, "node_id", nodeID, "user_id", userID)
		// Don't return error - XP award failure should not block node completion
	} else {
		s.logger.InfoContext(ctx, "XP awarded for node completion", "node_id", nodeID, "user_id", userID, "xp", repository.XPPerNodeComplete)
	}

	utils.SafeGo(ctx, s.logger, "Mission_CompleteNode", 10*time.Second, func(bgCtx context.Context) error {
		return s.missionSvc.ProcessMissionEvent(bgCtx, userID, missionModel.ConditionCompleteNode)
	})

	// Fetch node details to get pathID
	node, err := s.nodeRepo.GetNodeByID(ctx, nodeID, userID)
	if err == nil && node != nil {
		// Calculate the parent path's new completion state
		progress, err := s.progressRepo.GetUserPathProgress(ctx, node.PathID, userID)
		if err == nil && progress != nil {
			if progress.TotalNodes > 0 && progress.CompletedNodes >= progress.TotalNodes {
				// Mark path enrollment fully completed in database
				if err := s.pathRepo.UpdatePathEnrollmentCompletion(ctx, node.PathID, userID); err != nil {
					s.logger.ErrorContext(ctx, "failed to update path enrollment completion", "error", err, "path_id", node.PathID, "user_id", userID)
				} else {
					utils.SafeGo(ctx, s.logger, "Mission_CompletePath", 10*time.Second, func(bgCtx context.Context) error {
						return s.missionSvc.ProcessMissionEvent(bgCtx, userID, missionModel.ConditionCompletePath)
					})
					s.logger.InfoContext(ctx, "path fully completed", "path_id", node.PathID, "user_id", userID)

					// Award bonus XP for completing entire path
					if err := s.xpRepo.AddXPAndRecalcLevel(ctx, userID, repository.XPPerPathComplete); err != nil {
						s.logger.ErrorContext(ctx, "failed to award bonus XP for path completion", "error", err, "path_id", node.PathID, "user_id", userID)
					} else {
						s.logger.InfoContext(ctx, "bonus XP awarded for path completion", "path_id", node.PathID, "user_id", userID, "xp", repository.XPPerPathComplete)
					}
				}
			}
		}
	}

	s.logger.InfoContext(ctx, "node completed successfully", "node_id", nodeID, "user_id", userID)
	return nil
}
