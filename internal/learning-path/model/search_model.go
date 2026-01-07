package model

import "time"

// SearchPathRequest represents the request for searching learning paths
type SearchPathRequest struct {
	Query   string                 `json:"query" binding:"required"`
	TopK    int                    `json:"top_k"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// SearchPathResult represents a single search result with full path details
type SearchPathResult struct {
	PathID      string    `json:"path_id"`
	Score       float64   `json:"score"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CoverImgURL string    `json:"cover_img_url,omitempty"`
	Objective   string    `json:"objective,omitempty"`
	AvgRating   float64   `json:"avg_rating"`
	Status      string    `json:"status"`
	CreatorID   string    `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SearchPathResponse represents the response for search results
type SearchPathResponse struct {
	Query   string             `json:"query"`
	Total   int                `json:"total"`
	Results []SearchPathResult `json:"results"`
}
