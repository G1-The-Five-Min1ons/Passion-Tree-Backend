package model

type NodeProgressDetail struct {
	NodeID string `json:"node_id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type PathProgressResponse struct {
	PathID         string               `json:"path_id"`
	UserID         string               `json:"user_id"`
	TotalNodes     int                  `json:"total_nodes"`
	CompletedNodes int                  `json:"completed_nodes"`
	Progress       float64              `json:"progress_percentage"`
}