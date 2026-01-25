package model

type HistoryResponse struct {
	Target_path    string     `json:"target_path"`
	Path_id      string     `json:"path_id"`
}

type GetHistoryRequest struct {
	UserID string `json:"user_id" binding:"required"`
}