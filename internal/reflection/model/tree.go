package model

import "time"

type Tree struct {
	TreeID       string    `json:"tree_id"`
	Title        string    `json:"title"`
	Difficulties string    `json:"difficulties"`
	Status       string    `json:"status"`
	IsPause      bool      `json:"is_pause"`
	NodeCount    int       `json:"node_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastUpdate   time.Time `json:"last_update"`
	AlbumID      string    `json:"album_id"`
	PathID       string    `json:"path_id"`
}

// CreateTreeRequest
type CreateTreeRequest struct {
	Title        string `json:"title" binding:"required"`
	Difficulties string `json:"difficulties" binding:"required"`
	PathID       string `json:"path_id" binding:"required"`
	AlbumID      string `json:"album_id" binding:"required"`
}

// UpdateTreeRequest
type UpdateTreeRequest struct {
	Title   string `json:"title"`
	AlbumID string `json:"album_id,omitempty"`
	// Note: Use PATCH /trees/:tree_id/pause to toggle pause status
}

// TreeResponse
type TreeResponse struct {
	TreeID       string                   `json:"tree_id"`
	Title        string                   `json:"title"`
	Difficulties string                   `json:"difficulties"`
	Status       string                   `json:"status"`
	IsPause      bool                     `json:"is_pause"`
	NodeCount    int                      `json:"node_count"`
	CreatedAt    time.Time                `json:"created_at"`
	LastUpdate   time.Time                `json:"last_update"`
	AlbumID      string                   `json:"album_id"`
	PathID       string                   `json:"path_id"`
	Nodes        []TreeNode               `json:"nodes,omitempty"`
}

// PauseTreeRequest - body is optional, pause status will be toggled automatically
type PauseTreeRequest struct {
	// Deprecated: This field is no longer used. The endpoint toggles pause status automatically.
	IsPause *bool `json:"is_pause,omitempty"`
}
