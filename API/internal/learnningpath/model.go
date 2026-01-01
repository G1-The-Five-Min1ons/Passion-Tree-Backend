package learningpath

import "time"

type LearningPath struct {
	PathID      string    `json:"path_id"`
	Title       string    `json:"title"`
	CoverImgURL string    `json:"cover_img_url"`
	Objective   string    `json:"objective"`
	Description string    `json:"description"`
	AvgRating   float64   `json:"avg_rating"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"create_at"`
	UpdatedAt   time.Time `json:"update_at"`
	Nodes       []Node    `json:"nodes,omitempty"`
}

type Node struct {
	NodeID      string `json:"node_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type PathEnroll struct {
	EnrollID   string     `json:"enroll_id"`
	Status     string     `json:"status"`
	EnrollAt   time.Time  `json:"enroll_at"`
	CompleteAt *time.Time `json:"complete_at"`
}

type NodeProgress struct {
	Node
	Status string `json:"status"`
}

type CreatePathRequest struct {
	Title       string `json:"title" binding:"required"`
	Objective   string `json:"objective"`
	Description string `json:"description"`
	CoverImgURL string `json:"cover_img_url"`
	Status      string `json:"status"`
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