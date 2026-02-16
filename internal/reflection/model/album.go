package model

import "time"

// Album represents a tree album in the database
type Album struct {
	AlbumID       string    `json:"album_id"`
	AlbumName     string    `json:"album_name"`
	TreeCount     int       `json:"tree_count"`
	CoverImageURL string    `json:"cover_image_url"`
	CreatedAt     time.Time `json:"created_at"`
	LastEdit      time.Time `json:"last_edit"`
	UserID        string    `json:"user_id"`
}

// CreateAlbumRequest represents the request to create an album
type CreateAlbumRequest struct {
	UserID       string `json:"user_id"`
	AlbumName    string `json:"album_name"`
	CoverImageURL  string `json:"cover_image_url"`
}

// UpdateAlbumRequest represents the request to update an album
type UpdateAlbumRequest struct {
	AlbumName    string `json:"album_name"`
	CoverImageURL  string `json:"cover_image_url"`
}

// AlbumResponse represents the response after creating/updating an album
type AlbumResponse struct {
	AlbumID       string    `json:"album_id"`
	AlbumName     string    `json:"album_name"`
	TreeCount     int       `json:"tree_count"`
	CoverImageURL string    `json:"cover_image_url"`
	UserID        string    `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	LastEdit      time.Time `json:"last_edit"`
}
