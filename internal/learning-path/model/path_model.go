package model

import "time"

type CreatePathRequest struct {
	Title       string `json:"title" binding:"required"`
	Objective   string `json:"objective"`
	Description string `json:"description"`
	CoverImgURL string `json:"cover_img_url"`
	Status      string `json:"status"`
	CreatorID   string `json:"creator_id"`
}

type UpdatePathRequest struct {
	Title       string `json:"title"`
	Objective   string `json:"objective"`
	Description string `json:"description"`
	CoverImgURL string `json:"cover_img_url"`
	Status      string `json:"status"`
}

type StartPathRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type LearningPath struct {
	PathID      string    `json:"path_id"`
	Title       string    `json:"title"`
	CoverImgURL string    `json:"cover_img_url"`
	Objective   string    `json:"objective"`
	Description string    `json:"description"`
	AvgRating   float64   `json:"avg_rating"`
	Status      string    `json:"status"`
	CreatorID   string    `json:"creator_id"`
	CreatedAt   time.Time `json:"create_at"`
	UpdatedAt   time.Time `json:"update_at"`
	Nodes       []Node    `json:"nodes,omitempty"`
}

type PathEnroll struct {
	EnrollID   string     `json:"enroll_id"`
	Status     string     `json:"status"`
	EnrollAt   time.Time  `json:"enroll_at"`
	CompleteAt *time.Time `json:"complete_at"`
}