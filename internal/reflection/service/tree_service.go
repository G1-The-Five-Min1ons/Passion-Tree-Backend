package service

import (
	"context"
	"database/sql"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

func (s *serviceImpl) CreateTree(ctx context.Context, req model.CreateTreeRequest) (*model.TreeResponse, error) {
	s.logger.InfoContext(ctx, "creating new tree", "album_id", req.AlbumID, "title", req.Title)

	if req.Title == "" {
		return nil, apperror.NewBadRequest("title is required")
	}
	if req.AlbumID == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}
	
	treeID, err := s.refRepo.CreateTree(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create tree in database", "error", err, "album_id", req.AlbumID)
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid album_id: album does not exist")
		}
		return nil, apperror.NewInternal(err)
	}
	
	// Get the created tree to return full details
	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tree after creation", "error", err, "tree_id", treeID)
		return nil, apperror.NewInternal(err)
	}

	s.logger.InfoContext(ctx, "tree created successfully", "tree_id", treeID)
	
	return &model.TreeResponse{
		TreeID:  tree.TreeID,
		Title:   tree.Title,
		Status:  tree.Status,
		IsPause: tree.IsPause,
		AlbumID: tree.AlbumID,
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

		return nil, apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "successfully retrieved tree", "tree_id", treeID)
	return tree, nil
}

func (s *serviceImpl) GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error) {
	s.logger.InfoContext(ctx, "fetching trees for album", "album_id", albumID)

	if albumID == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}
	
	trees, err := s.refRepo.GetTreesByAlbumID(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "no trees found for album", "album_id", albumID)
			return nil, apperror.NewNotFound("trees for album with id '%s' not found", albumID)
		}
		s.logger.ErrorContext(ctx, "database error fetching album trees", "error", err, "album_id", albumID)
		return nil, apperror.NewInternal(err)
	}
	
	if len(trees) == 0 {
		s.logger.InfoContext(ctx, "album has an empty tree list", "album_id", albumID)
		return nil, apperror.NewNotFound("trees for album with id '%s' not found", albumID)
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
		return apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "tree updated successfully", "tree_id", treeID)
	return nil
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
		return apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "tree deleted successfully", "tree_id", treeID)
	return nil
}
