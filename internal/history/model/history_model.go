package model

import "time"

type HistoryResponse struct {
	EnrollID    string     `json:"enroll_id"`
	Status      string     `json:"status"`
	EnrollAt    time.Time  `json:"enroll_at"`
	CompleteAt  *time.Time `json:"complete_at"`
	PathID      string     `json:"path_id"`
	Title       string     `json:"title"`
	CoverImgURL string     `json:"cover_img_url"`
	Objective   string     `json:"objective"`
	CreatorID   string     `json:"creator_id"`
}

type GetHistoryRequest struct {
	UserID string `json:"user_id" binding:"required"`
}