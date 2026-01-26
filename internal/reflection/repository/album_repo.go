package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/reflection/model"

	"github.com/google/uuid"
)

// CreateAlbum creates a new album in the database
func (r *repositoryImpl) CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (string, error) {
	albumID := uuid.New().String()
	
	query := `
		INSERT INTO tree_album (album_id, album_name, tree_count, cover_image_url, create_at, last_edit, user_id)
		VALUES (@p1, @p2, 0, @p3, GETDATE(), GETDATE(), @p4)
	`
	
	_, err := r.db.ExecContext(ctx, query, albumID, req.AlbumName, req.CoverImgURL, req.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to create album: %w", err)
	}
	
	return albumID, nil
}

// GetAlbumByID retrieves an album by its ID
func (r *repositoryImpl) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), album_id) as album_id, album_name, tree_count, cover_image_url, create_at, last_edit, CONVERT(VARCHAR(36), user_id) as user_id
		FROM tree_album
		WHERE album_id = @p1
	`
	
	var album model.Album
	err := r.db.QueryRowContext(ctx, query, albumID).Scan(
		&album.AlbumID,
		&album.AlbumName,
		&album.TreeCount,
		&album.CoverImageURL,
		&album.CreatedAt,
		&album.LastEdit,
		&album.UserID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetAlbumByID scan failed: %w", err)
	}
	
	return &album, nil
}

// GetAlbumsByUserID retrieves all albums for a specific user
func (r *repositoryImpl) GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), album_id) as album_id, album_name, tree_count, cover_image_url, create_at, last_edit, CONVERT(VARCHAR(36), user_id) as user_id
		FROM tree_album
		WHERE user_id = @p1
		ORDER BY last_edit DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %w", err)
	}
	defer rows.Close()
	
	var albums []model.Album
	for rows.Next() {
		var album model.Album
		err := rows.Scan(
			&album.AlbumID,
			&album.AlbumName,
			&album.TreeCount,
			&album.CoverImageURL,
			&album.CreatedAt,
			&album.LastEdit,
			&album.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan album: %w", err)
		}
		albums = append(albums, album)
	}
	
	return albums, nil
}

// UpdateAlbum updates an existing album
func (r *repositoryImpl) UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
	query := `
		UPDATE tree_album
		SET album_name = @p1, cover_image_url = @p2, last_edit = GETDATE()
		WHERE album_id = @p3
	`
	
	result, err := r.db.ExecContext(ctx, query, req.AlbumName, req.CoverImageURL, albumID)
	if err != nil {
		return fmt.Errorf("failed to update album: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// DeleteAlbum deletes an album by its ID
func (r *repositoryImpl) DeleteAlbum(ctx context.Context, albumID string) error {
	query := `DELETE FROM tree_album WHERE album_id = @p1`
	
	result, err := r.db.ExecContext(ctx, query, albumID)
	if err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}
