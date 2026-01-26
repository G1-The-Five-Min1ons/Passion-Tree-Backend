package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

func (s *serviceImpl) CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (*model.AlbumResponse, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}
	if strings.TrimSpace(req.AlbumName) == "" {
		return nil, apperror.NewBadRequest("album_name is required")
	}
	
	albumID, err := s.refRepo.CreateAlbum(ctx, req)
	if err != nil {
		fmt.Printf("CreateAlbum database error: %v\n", err)
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid user_id: user does not exist")
		}
		return nil, apperror.NewInternal(err)
	}
	
	// Get the created album to return full details
	album, err := s.refRepo.GetAlbumByID(ctx, albumID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	
	return &model.AlbumResponse{
		AlbumID:       album.AlbumID,
		AlbumName:     album.AlbumName,
		TreeCount:     album.TreeCount,
		CoverImageURL: album.CoverImageURL,
	}, nil
}

func (s *serviceImpl) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	if strings.TrimSpace(albumID) == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}
	
	album, err := s.refRepo.GetAlbumByID(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("album with id '%s' not found", albumID)
		}
		return nil, apperror.NewInternal(err)
	}
	
	return album, nil
}

func (s *serviceImpl) GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}
	
	albums, err := s.refRepo.GetAlbumsByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	
	return albums, nil
}

func (s *serviceImpl) UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
	if strings.TrimSpace(albumID) == "" {
		return apperror.NewBadRequest("album_id is required")
	}
	if strings.TrimSpace(req.AlbumName) == "" {
		return apperror.NewBadRequest("album_name is required")
	}
	
	err := s.refRepo.UpdateAlbum(ctx, albumID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("album with id '%s' not found", albumID)
		}
		return apperror.NewInternal(err)
	}
	
	return nil
}

func (s *serviceImpl) DeleteAlbum(ctx context.Context, albumID string) error {
	if strings.TrimSpace(albumID) == "" {
		return apperror.NewBadRequest("album_id is required")
	}
	
	err := s.refRepo.DeleteAlbum(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("album with id '%s' not found", albumID)
		}
		return apperror.NewInternal(err)
	}
	
	return nil
}
