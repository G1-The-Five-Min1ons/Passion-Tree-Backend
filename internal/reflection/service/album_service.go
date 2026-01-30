package service

import (
	"context"
	"database/sql"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

func (s *serviceImpl) CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (*model.AlbumResponse, error) {
	s.logger.InfoContext(ctx, "creating new album", "user_id", req.UserID, "album_name", req.AlbumName)
	
	if req.UserID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}
	if req.AlbumName == "" {
		return nil, apperror.NewBadRequest("album_name is required")
	}
	
	albumID, err := s.refRepo.CreateAlbum(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create album", "error", err, "user_id", req.UserID, "album_name", req.AlbumName)
		if apperror.IsForeignKeyError(err) {
			return nil, apperror.NewBadRequest("invalid user_id: user does not exist")
		}
		return nil, apperror.NewInternal(err)
	}
	
	album, err := s.refRepo.GetAlbumByID(ctx, albumID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get album after creation", "error", err, "album_id", albumID)
		return nil, apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "album created successfully", "album_id", albumID)
	
	return &model.AlbumResponse{
		AlbumID:       album.AlbumID,
		AlbumName:     album.AlbumName,
		TreeCount:     album.TreeCount,
		CoverImageURL: album.CoverImageURL,
	}, nil
}

func (s *serviceImpl) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	if albumID == "" {
		return nil, apperror.NewBadRequest("album_id is required")
	}
	
	album, err := s.refRepo.GetAlbumByID(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "album not found", "album_id", albumID)
			return nil, apperror.NewNotFound("album with id '%s' not found", albumID)
		}
		s.logger.ErrorContext(ctx, "database error fetching album", "error", err, "album_id", albumID)
		return nil, apperror.NewInternal(err)
	}
	
	return album, nil
}

func (s *serviceImpl) GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error) {
	s.logger.InfoContext(ctx, "fetching albums for user", "user_id", userID)
	
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}
	
	albums, err := s.refRepo.GetAlbumsByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "no albums found for user", "user_id", userID)
			return nil, apperror.NewNotFound("albums for user with id '%s' not found", userID)
		}
		s.logger.ErrorContext(ctx, "failed to fetch user albums", "error", err, "user_id", userID)
		return nil, apperror.NewInternal(err)
	}
	
	if len(albums) == 0 {
		s.logger.InfoContext(ctx, "user has an empty album list", "user_id", userID)
		return nil, apperror.NewNotFound("albums for user with id '%s' not found", userID)
	}

	s.logger.InfoContext(ctx, "successfully retrieved user albums", "user_id", userID, "count", len(albums))
	return albums, nil
}

func (s *serviceImpl) UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
	s.logger.InfoContext(ctx, "updating album", "album_id", albumID, "album_name", req.AlbumName)
	
	if albumID == "" {
		return apperror.NewBadRequest("album_id is required")
	}
	if req.AlbumName == "" {
		return apperror.NewBadRequest("album_name is required")
	}
	
	err := s.refRepo.UpdateAlbum(ctx, albumID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: album not found", "album_id", albumID)
			return apperror.NewNotFound("album with id '%s' not found", albumID)
		}
		s.logger.ErrorContext(ctx, "failed to update album", "error", err, "album_id", albumID)
		return apperror.NewInternal(err)
	}
	
	s.logger.InfoContext(ctx, "album updated successfully", "album_id", albumID)

	return nil
}

func (s *serviceImpl) DeleteAlbum(ctx context.Context, albumID string) error {
	s.logger.InfoContext(ctx, "request to delete album", "album_id", albumID)

	if albumID == "" {
		return apperror.NewBadRequest("album_id is required")
	}
	
	err := s.refRepo.DeleteAlbum(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("album with id '%s' not found", albumID)
		}
		s.logger.ErrorContext(ctx, "failed to delete album", "error", err, "album_id", albumID)
		return apperror.NewInternal(err)
	}
	
	return nil
}
