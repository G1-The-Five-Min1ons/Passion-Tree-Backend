package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

func (s *serviceImpl) CreateTree(ctx context.Context, req model.CreateTreeRequest) (*model.TreeResponse, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, apperror.NewBadRequest("title is required")
	}
	if strings.TrimSpace(req.AlbumID) == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}
	
	treeID, err := s.refRepo.CreateTree(ctx, req)
	if err != nil {
		fmt.Printf("CreateTree database error: %v\n", err)
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid album_id: album does not exist")
		}
		return nil, apperror.NewInternal(err)
	}
	
	// Get the created tree to return full details
	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	
	return &model.TreeResponse{
		TreeID:  tree.TreeID,
		Title:   tree.Title,
		Status:  tree.Status,
		IsPause: tree.IsPause,
		AlbumID: tree.AlbumID,
	}, nil
}

func (s *serviceImpl) GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error) {
	if strings.TrimSpace(treeID) == "" {
		return nil, apperror.NewBadRequest("tree_id is required")
	}
	
	tree, err := s.refRepo.GetTreeByID(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		return nil, apperror.NewInternal(err)
	}
	
	return tree, nil
}

func (s *serviceImpl) GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error) {
	if strings.TrimSpace(albumID) == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}
	
	trees, err := s.refRepo.GetTreesByAlbumID(ctx, albumID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	
	return trees, nil
}

func (s *serviceImpl) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
	if strings.TrimSpace(treeID) == "" {
		return apperror.NewBadRequest("tree_id is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return apperror.NewBadRequest("title is required")
	}
	
	err := s.refRepo.UpdateTree(ctx, treeID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		return apperror.NewInternal(err)
	}
	
	return nil
}

func (s *serviceImpl) DeleteTree(ctx context.Context, treeID string) error {
	if strings.TrimSpace(treeID) == "" {
		return apperror.NewBadRequest("tree_id is required")
	}
	
	err := s.refRepo.DeleteTree(ctx, treeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("tree with id '%s' not found", treeID)
		}
		return apperror.NewInternal(err)
	}
	
	return nil
}
