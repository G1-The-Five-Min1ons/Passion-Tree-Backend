package service

import (
	"context"
	"database/sql"
	"errors"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
	"time"
)

func (s *serviceImpl) CreateTree(ctx context.Context, req model.CreateTreeRequest) (*model.TreeResponse, error) {
	s.logger.InfoContext(ctx, "creating new tree", "album_id", req.AlbumID, "title", req.Title, "path_id", req.PathID)

	// Validate required fields
	if req.Title == "" {
		return nil, apperror.NewBadRequest("title is required")
	}
	if req.Difficulties == "" {
		return nil, apperror.NewBadRequest("difficulties is required")
	}
	if req.PathID == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}
	if req.AlbumID == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}

	// Create the tree (this will also create tree_node records)
	treeID, err := s.refRepo.CreateTree(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create tree in database", "error", err, "album_id", req.AlbumID)
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid album_id or path_id: referenced resource does not exist")
		}
		return nil, apperror.NewInternal("failed to create tree: %w", err)
	}

	// Get the created tree to return full details
	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tree after creation", "error", err, "tree_id", treeID)
		return nil, apperror.NewInternal("failed to get tree after creation: %w", err)
	}

	// Fetch tree nodes
	nodes, err := s.refRepo.GetTreeNodesByTreeID(ctx, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch tree nodes", "error", err, "tree_id", treeID)
		return nil, apperror.NewInternal("%s", err.Error())
	}

	s.logger.InfoContext(ctx, "tree created successfully", "tree_id", treeID, "nodes_count", len(nodes))

	return &model.TreeResponse{
		TreeID:        tree.TreeID,
		Title:         tree.Title,
		Difficulties:  tree.Difficulties,
		Status:        normalizeTreeStatus(computeTreeStatus(tree.Difficulties, tree.LastReflectAt, tree.IsPause, tree.PausedAt)),
		IsPause:       tree.IsPause,
		NodeCount:     tree.NodeCount,
		CreatedAt:     tree.CreatedAt,
		LastUpdate:    tree.LastUpdate,
		AlbumID:       tree.AlbumID,
		PathID:        tree.PathID,
		LastReflectAt: tree.LastReflectAt,
		TreeScore:     tree.TreeScore,
		Nodes:         nodes,
	}, nil
}

func (s *serviceImpl) GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error) {
	s.logger.InfoContext(ctx, "fetching tree by ID", "tree_id", treeID)

	if treeID == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}

	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "tree not found", "tree_id", treeID)
			return nil, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		s.logger.ErrorContext(ctx, "database error fetching tree", "error", err, "tree_id", treeID)

		return nil, apperror.NewInternal("database error fetching tree: %w", err)
	}

	// Compute live status and persist it back when the stored value is stale.
	tree.Status = s.syncTreeStatus(
		ctx,
		tree.TreeID,
		tree.Status,
		tree.Difficulties,
		tree.LastReflectAt,
		tree.IsPause,
		tree.PausedAt,
	)

	s.logger.InfoContext(ctx, "successfully retrieved tree", "tree_id", treeID)
	return tree, nil
}

func (s *serviceImpl) GetTreesByAlbumID(ctx context.Context, albumID string, includeNodes bool, userID string) (interface{}, error) {
	s.logger.InfoContext(ctx, "fetching trees for album", "album_id", albumID, "include_nodes", includeNodes, "user_id", userID)

	if albumID == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}

	// If nodes are requested, use optimized query
	if includeNodes {
		treesWithNodes, err := s.refRepo.GetTreesWithNodesByAlbumID(ctx, albumID, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				s.logger.WarnContext(ctx, "no trees found for album", "album_id", albumID)
				return []model.TreeResponse{}, nil
			}
			s.logger.ErrorContext(ctx, "database error fetching album trees with nodes", "error", err, "album_id", albumID)
			return nil, apperror.NewInternal("database error fetching album trees: %w", err)
		}

		if len(treesWithNodes) == 0 {
			s.logger.InfoContext(ctx, "album has an empty tree list", "album_id", albumID)
			return []model.TreeResponse{}, nil
		}

		// Compute live status for each tree and persist stale status values.
		for i := range treesWithNodes {
			treesWithNodes[i].Status = s.syncTreeStatus(
				ctx,
				treesWithNodes[i].TreeID,
				treesWithNodes[i].Status,
				treesWithNodes[i].Difficulties,
				treesWithNodes[i].LastReflectAt,
				treesWithNodes[i].IsPause,
				treesWithNodes[i].PausedAt,
			)
		}

		s.logger.InfoContext(ctx, "successfully retrieved album trees with nodes", "album_id", albumID, "count", len(treesWithNodes))
		return treesWithNodes, nil
	}

	// Default: return trees without nodes
	trees, err := s.refRepo.GetTreesByAlbumID(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "no trees found for album", "album_id", albumID)
			return []model.Tree{}, nil
		}
		s.logger.ErrorContext(ctx, "database error fetching album trees", "error", err, "album_id", albumID)
		return nil, apperror.NewInternal("database error fetching album trees: %w", err)
	}

	if len(trees) == 0 {
		s.logger.InfoContext(ctx, "album has an empty tree list", "album_id", albumID)
		return []model.Tree{}, nil
	}

	// Compute live status for each tree and persist stale status values.
	for i := range trees {
		trees[i].Status = s.syncTreeStatus(
			ctx,
			trees[i].TreeID,
			trees[i].Status,
			trees[i].Difficulties,
			trees[i].LastReflectAt,
			trees[i].IsPause,
			trees[i].PausedAt,
		)
	}

	s.logger.InfoContext(ctx, "successfully retrieved album trees", "album_id", albumID, "count", len(trees))
	return trees, nil
}

func (s *serviceImpl) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
	s.logger.InfoContext(ctx, "updating tree", "tree_id", treeID, "title", req.Title)

	if treeID == "" {
		return apperror.NewBadRequest("tree_id is required")
	}
	if req.Title == "" {
		return apperror.NewBadRequest("title is required")
	}

	err := s.refRepo.UpdateTree(ctx, treeID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: tree not found", "tree_id", treeID)
			return apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		s.logger.ErrorContext(ctx, "failed to update tree", "error", err, "tree_id", treeID)
		return apperror.NewInternal("failed to update tree: %w", err)
	}

	s.logger.InfoContext(ctx, "tree updated successfully", "tree_id", treeID)
	return nil
}

func (s *serviceImpl) RetrieveTree(ctx context.Context, treeID string, userID string) (*model.RetrieveTreeResponse, error) {
	s.logger.InfoContext(ctx, "request to retrieve tree", "tree_id", treeID, "user_id", userID)

	if treeID == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}
	if userID == "" {
		return nil, apperror.NewUnauthorized("user not authenticated")
	}

	const retrieveCost = 5

	remainingHearts, err := s.refRepo.RetrieveTree(ctx, treeID, userID, retrieveCost)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInsufficientHearts):
			s.logger.WarnContext(ctx, "retrieve failed due to insufficient hearts", "tree_id", treeID, "user_id", userID)
			return nil, apperror.NewBadRequest("insufficient hearts: at least 5 hearts are required")
		case err == sql.ErrNoRows:
			s.logger.WarnContext(ctx, "retrieve failed: tree not found or access denied", "tree_id", treeID, "user_id", userID)
			return nil, apperror.NewNotFound("tree with id '%s' not found", treeID)
		default:
			s.logger.ErrorContext(ctx, "failed to retrieve tree", "tree_id", treeID, "user_id", userID, "error", err)
			return nil, apperror.NewInternal("failed to retrieve tree: %w", err)
		}
	}

	now := time.Now()
	response := &model.RetrieveTreeResponse{
		TreeID:        treeID,
		HeartCount:    remainingHearts,
		Status:        statusGrowing,
		LastReflectAt: &now,
	}

	s.logger.InfoContext(ctx, "tree retrieved successfully", "tree_id", treeID, "user_id", userID, "remaining_hearts", remainingHearts)
	return response, nil
}

func (s *serviceImpl) DeleteTree(ctx context.Context, treeID string) error {
	s.logger.InfoContext(ctx, "request to delete tree", "tree_id", treeID)

	if treeID == "" {
		return apperror.NewBadRequest("tree_id is required")
	}

	err := s.refRepo.DeleteTree(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		s.logger.ErrorContext(ctx, "failed to delete tree", "error", err, "tree_id", treeID)
		return apperror.NewInternal("failed to delete tree: %w", err)
	}

	s.logger.InfoContext(ctx, "tree deleted successfully", "tree_id", treeID)
	return nil
}

func (s *serviceImpl) PauseTree(ctx context.Context, treeID string, req model.PauseTreeRequest) (bool, error) {
	s.logger.InfoContext(ctx, "request to toggle pause/unpause tree", "tree_id", treeID)

	if treeID == "" {
		return false, apperror.NewBadRequest("tree_id is required")
	}

	// Get current tree state
	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "tree not found", "tree_id", treeID)
			return false, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		s.logger.ErrorContext(ctx, "database error fetching tree", "error", err, "tree_id", treeID)
		return false, apperror.NewInternal("%s", err.Error())
	}

	// Toggle pause status (pause -> unpause, unpause -> pause)
	newPauseStatus := !tree.IsPause

	// Update pause status
	err = s.refRepo.PauseTree(ctx, treeID, newPauseStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		s.logger.ErrorContext(ctx, "failed to pause/unpause tree", "error", err, "tree_id", treeID)
		return false, apperror.NewInternal("%s", err.Error())
	}

	updatedTree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to refresh tree after pause toggle", "tree_id", treeID, "error", err)
	} else {
		updatedTree.Status = s.syncTreeStatus(
			ctx,
			updatedTree.TreeID,
			updatedTree.Status,
			updatedTree.Difficulties,
			updatedTree.LastReflectAt,
			updatedTree.IsPause,
			updatedTree.PausedAt,
		)
	}

	pauseStatus := "paused"
	if !newPauseStatus {
		pauseStatus = "unpaused"
	}
	s.logger.InfoContext(ctx, "tree "+pauseStatus+" successfully", "tree_id", treeID, "previous_status", tree.IsPause, "new_status", newPauseStatus)
	return newPauseStatus, nil
}

// CalculateAndUpdateTreeScore computes the average weighted_reflection_score
// across all reflected nodes in the tree on a 0-10 scale,
// and persists it to tree.tree_score.
func (s *serviceImpl) CalculateAndUpdateTreeScore(ctx context.Context, treeID string) (*float64, error) {
	s.logger.InfoContext(ctx, "calculating tree score", "tree_id", treeID)

	if treeID == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}

	score, err := s.refRepo.CalculateAndUpdateTreeScore(ctx, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to calculate tree score", "tree_id", treeID, "error", err)
		return nil, apperror.NewInternal("failed to calculate tree score: %w", err)
	}

	s.logger.InfoContext(ctx, "tree score calculated and saved", "tree_id", treeID, "score", score)
	return score, nil
}
