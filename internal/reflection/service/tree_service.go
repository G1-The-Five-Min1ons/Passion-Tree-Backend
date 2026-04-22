package service

import (
	"context"
	"database/sql"
	"errors"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
	"strings"
	"time"
)

func (s *serviceImpl) CreateTree(ctx context.Context, req model.CreateTreeRequest, userID string) (*model.TreeResponse, error) {
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
	if userID == "" {
		return nil, apperror.NewUnauthorized("user not authenticated")
	}

	if err := s.ensureAlbumOwnership(ctx, req.AlbumID, userID); err != nil {
		return nil, err
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
	if tree == nil {
		return nil, apperror.NewInternal("failed to get tree after creation: empty result")
	}

	// Fetch tree nodes
	nodes, err := s.refRepo.GetTreeNodesByTreeID(ctx, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch tree nodes", "error", err, "tree_id", treeID)
		return nil, apperror.NewInternal("failed to fetch tree nodes: %w", err)
	}

	s.logger.InfoContext(ctx, "tree created successfully", "tree_id", treeID, "nodes_count", len(nodes))

	return &model.TreeResponse{
		TreeID:             tree.TreeID,
		Title:              tree.Title,
		Difficulties:       tree.Difficulties,
		Status:             normalizeTreeStatus(computeTreeStatus(tree.Difficulties, tree.LastReflectAt, tree.IsPause, tree.PausedAt)),
		IsPause:            tree.IsPause,
		IsReflectionClosed: tree.IsReflectionClosed,
		NodeCount:          tree.NodeCount,
		CreatedAt:          tree.CreatedAt,
		LastUpdate:         tree.LastUpdate,
		AlbumID:            tree.AlbumID,
		PathID:             tree.PathID,
		LastReflectAt:      tree.LastReflectAt,
		PauseFrom:          tree.PauseFrom,
		PauseTo:            tree.PauseTo,
		PausedAt:           tree.PausedAt,
		TreeScore:          tree.TreeScore,
		Nodes:              nodes,
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
	if tree == nil {
		return nil, apperror.NewInternal("database error fetching tree: empty result")
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
		tree.IsReflectionClosed,
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
				treesWithNodes[i].IsReflectionClosed,
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
			trees[i].IsReflectionClosed,
		)
	}

	s.logger.InfoContext(ctx, "successfully retrieved album trees", "album_id", albumID, "count", len(trees))
	return trees, nil
}

func (s *serviceImpl) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest, userID string) error {
	s.logger.InfoContext(ctx, "updating tree", "tree_id", treeID, "title", req.Title)

	if treeID == "" {
		return apperror.NewBadRequest("tree_id is required")
	}
	if userID == "" {
		return apperror.NewUnauthorized("user not authenticated")
	}
	if req.Title == "" {
		return apperror.NewBadRequest("title is required")
	}

	if err := s.ensureTreeOwnership(ctx, treeID, userID); err != nil {
		return err
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

	if err := s.ensureTreeOwnership(ctx, treeID, userID); err != nil {
		return nil, err
	}

	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "tree not found", "tree_id", treeID)
			return nil, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		return nil, apperror.NewInternal("failed to load tree state: %w", err)
	}
	if tree == nil {
		return nil, apperror.NewInternal("failed to load tree state: empty result")
	}
	if tree.IsReflectionClosed {
		return nil, apperror.NewForbidden("this reflection tree has already been ended")
	}

	const retrieveCost = 5

	remainingHearts, err := s.refRepo.RetrieveTree(ctx, treeID, userID, retrieveCost)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.logger.WarnContext(ctx, "retrieve failed due to insufficient hearts", "tree_id", treeID, "user_id", userID)
			return nil, apperror.NewBadRequest("insufficient hearts: at least 5 hearts are required")
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

func (s *serviceImpl) EndReflecting(ctx context.Context, treeID string, userID string) (*model.TreeResponse, error) {
	s.logger.InfoContext(ctx, "request to end reflecting", "tree_id", treeID, "user_id", userID)

	if treeID == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}
	if userID == "" {
		return nil, apperror.NewUnauthorized("user not authenticated")
	}

	if err := s.ensureTreeOwnership(ctx, treeID, userID); err != nil {
		return nil, err
	}

	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "tree not found", "tree_id", treeID)
			return nil, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		s.logger.ErrorContext(ctx, "database error fetching tree", "error", err, "tree_id", treeID)
		return nil, apperror.NewInternal("failed to load tree state: %w", err)
	}
	if tree == nil {
		return nil, apperror.NewInternal("failed to load tree state: empty result")
	}

	if !tree.IsReflectionClosed {
		if err := s.refRepo.CloseTreeReflection(ctx, treeID, tree.Status); err != nil {
			s.logger.ErrorContext(ctx, "failed to end reflecting", "tree_id", treeID, "error", err)
			return nil, apperror.NewInternal("failed to end reflecting: %w", err)
		}
	}

	updatedTree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		return nil, apperror.NewInternal("failed to reload tree after ending reflection: %w", err)
	}
	if updatedTree == nil {
		return nil, apperror.NewInternal("failed to reload tree after ending reflection: empty result")
	}

	return &model.TreeResponse{
		TreeID:             updatedTree.TreeID,
		Title:              updatedTree.Title,
		Difficulties:       updatedTree.Difficulties,
		Status:             updatedTree.Status,
		IsPause:            updatedTree.IsPause,
		IsReflectionClosed: updatedTree.IsReflectionClosed,
		NodeCount:          updatedTree.NodeCount,
		CreatedAt:          updatedTree.CreatedAt,
		LastUpdate:         updatedTree.LastUpdate,
		AlbumID:            updatedTree.AlbumID,
		PathID:             updatedTree.PathID,
		LastReflectAt:      updatedTree.LastReflectAt,
		PauseFrom:          updatedTree.PauseFrom,
		PauseTo:            updatedTree.PauseTo,
		PausedAt:           updatedTree.PausedAt,
		TreeScore:          updatedTree.TreeScore,
	}, nil
}

func (s *serviceImpl) DeleteTree(ctx context.Context, treeID string, userID string) error {
	s.logger.InfoContext(ctx, "request to delete tree", "tree_id", treeID)

	if treeID == "" {
		return apperror.NewBadRequest("tree_id is required")
	}
	if userID == "" {
		return apperror.NewUnauthorized("user not authenticated")
	}

	if err := s.ensureTreeOwnership(ctx, treeID, userID); err != nil {
		return err
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

func (s *serviceImpl) PauseTree(ctx context.Context, treeID string, userID string, req model.PauseTreeRequest) (*model.PauseTreeResponse, error) {
	s.logger.InfoContext(ctx, "request to pause tree", "tree_id", treeID, "user_id", userID)

	if treeID == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}
	if userID == "" {
		return nil, apperror.NewUnauthorized("user not authenticated")
	}

	if err := s.ensureTreeOwnership(ctx, treeID, userID); err != nil {
		return nil, err
	}

	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "tree not found", "tree_id", treeID)
			return nil, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		s.logger.ErrorContext(ctx, "database error fetching tree", "error", err, "tree_id", treeID)
		return nil, apperror.NewInternal("failed to load tree state: %w", err)
	}
	if tree == nil {
		return nil, apperror.NewInternal("failed to load tree state: empty result")
	}
	if tree.IsReflectionClosed {
		return nil, apperror.NewForbidden("this reflection tree has already been ended")
	}

	if tree.IsPause {
		if err := s.refRepo.DeactivateActivePauseWindow(ctx, treeID); err != nil {
			s.logger.ErrorContext(ctx, "failed to deactivate active pause window", "error", err, "tree_id", treeID, "user_id", userID)
			return nil, apperror.NewInternal("failed to resume tree: %w", err)
		}

		if err := s.refRepo.UnpauseTree(ctx, treeID, userID); err != nil {
			s.logger.ErrorContext(ctx, "failed to resume tree", "error", err, "tree_id", treeID, "user_id", userID)
			return nil, apperror.NewInternal("failed to resume tree: %w", err)
		}

		s.logger.InfoContext(ctx, "tree resumed successfully", "tree_id", treeID, "user_id", userID)
		return &model.PauseTreeResponse{
			TreeID:     treeID,
			IsPause:    false,
			HeartCount: 0,
			PausedAt:   nil,
		}, nil
	}

	if req.PausedAt == "" {
		return nil, apperror.NewBadRequest("paused_at is required")
	}

	resumeOn, err := time.Parse(time.RFC3339, req.PausedAt)
	if err != nil {
		return nil, apperror.NewBadRequest("paused_at must be a valid RFC3339 datetime")
	}

	pauseFrom := time.Now()
	if req.PauseFrom != "" {
		parsedPauseFrom, parseErr := time.Parse(time.RFC3339, req.PauseFrom)
		if parseErr != nil {
			return nil, apperror.NewBadRequest("pause_from must be a valid RFC3339 datetime")
		}
		pauseFrom = parsedPauseFrom
	}

	if !resumeOn.After(pauseFrom) {
		return nil, apperror.NewBadRequest("paused_at must be in the future")
	}

	remainingHearts, err := s.refRepo.PauseTree(ctx, treeID, userID, pauseFrom, resumeOn)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, apperror.NewBadRequest("insufficient hearts: at least 3 hearts are required")
		default:
			s.logger.ErrorContext(ctx, "failed to pause tree", "error", err, "tree_id", treeID, "user_id", userID)
			return nil, apperror.NewInternal("failed to pause tree: %w", err)
		}
	}

	s.logger.InfoContext(ctx, "tree paused successfully", "tree_id", treeID, "paused_at", resumeOn, "remaining_hearts", remainingHearts)

	return &model.PauseTreeResponse{
		TreeID:     treeID,
		IsPause:    true,
		HeartCount: remainingHearts,
		PausedAt:   &resumeOn,
	}, nil
}

func (s *serviceImpl) ensureTreeOwnership(ctx context.Context, treeID string, userID string) error {
	owned, err := s.refRepo.IsTreeOwnedByUser(ctx, treeID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to verify tree ownership", "tree_id", treeID, "user_id", userID, "error", err)
		return apperror.NewInternal("failed to verify tree ownership: %w", err)
	}

	if !owned {
		s.logger.WarnContext(ctx, "tree not found or access denied", "tree_id", treeID, "user_id", userID)
		return apperror.NewNotFound("tree with id '%s' not found", treeID)
	}

	return nil
}

func (s *serviceImpl) ensureAlbumOwnership(ctx context.Context, albumID string, userID string) error {
	album, err := s.refRepo.GetAlbumByID(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "album not found", "album_id", albumID, "user_id", userID)
			return apperror.NewNotFound("album with id '%s' not found", albumID)
		}
		s.logger.ErrorContext(ctx, "failed to verify album ownership", "album_id", albumID, "user_id", userID, "error", err)
		return apperror.NewInternal("failed to verify album ownership: %w", err)
	}

	if album == nil {
		return apperror.NewInternal("failed to verify album ownership: empty result")
	}

	if !strings.EqualFold(album.UserID, userID) {
		s.logger.WarnContext(ctx, "album not found or access denied", "album_id", albumID, "user_id", userID)
		return apperror.NewNotFound("album with id '%s' not found", albumID)
	}

	return nil
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
