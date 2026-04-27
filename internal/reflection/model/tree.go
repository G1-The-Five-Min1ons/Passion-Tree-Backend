package model

import (
	"time"
)

type Tree struct {
	TreeID             string     `json:"tree_id"`
	Title              string     `json:"title"`
	Difficulties       string     `json:"difficulties"`
	Status             string     `json:"status"`
	IsPause            bool       `json:"is_pause"`
	IsReflectionClosed bool       `json:"is_reflection_closed"`
	NodeCount          int        `json:"node_count"`
	CreatedAt          time.Time  `json:"created_at"`
	LastUpdate         time.Time  `json:"last_update"`
	AlbumID            string     `json:"album_id"`
	PathID             string     `json:"path_id"`
	LastReflectAt      *time.Time `json:"last_reflect_at"`
	PauseFrom          *time.Time `json:"pause_from,omitempty"`
	PauseTo            *time.Time `json:"pause_to,omitempty"`
	PausedAt           *time.Time `json:"paused_at,omitempty"`
	TreeScore          *float64   `json:"tree_score"`
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
}

// TreeResponse
type TreeResponse struct {
	TreeID             string     `json:"tree_id"`
	Title              string     `json:"title"`
	Difficulties       string     `json:"difficulties"`
	Status             string     `json:"status"`
	IsPause            bool       `json:"is_pause"`
	IsReflectionClosed bool       `json:"is_reflection_closed"`
	NodeCount          int        `json:"node_count"`
	CreatedAt          time.Time  `json:"created_at"`
	LastUpdate         time.Time  `json:"last_update"`
	AlbumID            string     `json:"album_id"`
	PathID             string     `json:"path_id"`
	LastReflectAt      *time.Time `json:"last_reflect_at"`
	PauseFrom          *time.Time `json:"pause_from,omitempty"`
	PauseTo            *time.Time `json:"pause_to,omitempty"`
	PausedAt           *time.Time `json:"paused_at,omitempty"`
	TreeScore          *float64   `json:"tree_score"`
	Nodes              []TreeNode `json:"nodes,omitempty"`
}

// PauseTreeRequest - body is optional, pause status will be toggled automatically
type PauseTreeRequest struct {
	PauseFrom string `json:"pause_from,omitempty"`
	PausedAt  string `json:"paused_at,omitempty"`
}

type PauseTreeResponse struct {
	TreeID     string     `json:"tree_id"`
	IsPause    bool       `json:"is_pause"`
	HeartCount int        `json:"heart_count"`
	PausedAt   *time.Time `json:"paused_at,omitempty"`
}

type RetrieveTreeResponse struct {
	TreeID        string     `json:"tree_id"`
	HeartCount    int        `json:"heart_count"`
	Status        string     `json:"status"`
	LastReflectAt *time.Time `json:"last_reflect_at"`
}
