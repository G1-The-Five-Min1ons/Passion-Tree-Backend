package model

import (
	"strings"
	"time"
)

// Tree status constants.
const (
	StatusGrowing = "growing"
	StatusFading  = "fading"
	StatusDying   = "dying"
	StatusDied    = "died"
)

func ComputeTreeStatus(difficulties string, lastReflectAt *time.Time, isPause bool, pausedAt *time.Time) string {
	if lastReflectAt == nil {
		return StatusGrowing // brand-new tree, not yet reflected
	}

	// While paused: freeze elapsed time at the moment pause started.
	reference := time.Now()
	if isPause && pausedAt != nil {
		reference = *pausedAt
	}

	elapsed := reference.Sub(*lastReflectAt)
	if elapsed < 0 {
		elapsed = 0
	}

	var fading, dying, died time.Duration
	switch strings.ToLower(difficulties) {
	case "easy":
		fading = 30 * 24 * time.Hour
		dying = 60 * 24 * time.Hour
		died = 90 * 24 * time.Hour
	case "medium":
		fading = 7 * 24 * time.Hour
		dying = 14 * 24 * time.Hour
		died = 21 * 24 * time.Hour
	case "hard":
		fading = 24 * time.Hour
		dying = 48 * time.Hour
		died = 72 * time.Hour
	default:
		return StatusGrowing
	}

	switch {
	case elapsed >= died:
		return StatusDied
	case elapsed >= dying:
		return StatusDying
	case elapsed >= fading:
		return StatusFading
	default:
		return StatusGrowing
	}
}

type Tree struct {
	TreeID        string     `json:"tree_id"`
	Title         string     `json:"title"`
	Difficulties  string     `json:"difficulties"`
	Status        string     `json:"status"`
	IsPause       bool       `json:"is_pause"`
	NodeCount     int        `json:"node_count"`
	CreatedAt     time.Time  `json:"created_at"`
	LastUpdate    time.Time  `json:"last_update"`
	AlbumID       string     `json:"album_id"`
	PathID        string     `json:"path_id"`
	LastReflectAt *time.Time `json:"last_reflect_at"`
	PausedAt      *time.Time `json:"paused_at,omitempty"`
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
	TreeID        string     `json:"tree_id"`
	Title         string     `json:"title"`
	Difficulties  string     `json:"difficulties"`
	Status        string     `json:"status"`
	IsPause       bool       `json:"is_pause"`
	NodeCount     int        `json:"node_count"`
	CreatedAt     time.Time  `json:"created_at"`
	LastUpdate    time.Time  `json:"last_update"`
	AlbumID       string     `json:"album_id"`
	PathID        string     `json:"path_id"`
	LastReflectAt *time.Time `json:"last_reflect_at"`
	Nodes         []TreeNode `json:"nodes,omitempty"`
}

// PauseTreeRequest - body is optional, pause status will be toggled automatically
type PauseTreeRequest struct {
	// Deprecated: This field is no longer used. The endpoint toggles pause status automatically.
	IsPause *bool `json:"is_pause,omitempty"`
}
