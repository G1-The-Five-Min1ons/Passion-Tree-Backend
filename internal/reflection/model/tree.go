package model

import "time"

type Tree struct {
	TreeID     string    `json:"tree_id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	IsPause    bool      `json:"is_pause"`
	CreatedAt  time.Time `json:"create_at"`
	LastUpdate time.Time `json:"last_update"`
	AlbumID    string    `json:"album_id"`
}

// CreateTreeReques
type CreateTreeRequest struct {
	Title   string `json:"title"`
	AlbumID string `json:"album_id"`
}

// UpdateTreeRequest
type UpdateTreeRequest struct {
	Title   string `json:"title"`
	Status  string `json:"status"`
	IsPause bool   `json:"is_pause"`
}

// TreeResponse
type TreeResponse struct {
	TreeID  string `json:"tree_id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	IsPause bool   `json:"is_pause"`
	AlbumID string `json:"album_id"`
}
