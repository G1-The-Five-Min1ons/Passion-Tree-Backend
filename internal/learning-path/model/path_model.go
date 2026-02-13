package model

import "time"

type CreatePathRequest struct {
	Title          string `json:"title" binding:"required"`
	Objective      string `json:"objective"`
	Description    string `json:"description"`
	CoverImgURL    string `json:"cover_img_url"`
	Publish_status string `json:"publish_status"`
	CreatorID      string `json:"creator_id"`
}

type UpdatePathRequest struct {
	Title          string `json:"title"`
	Objective      string `json:"objective"`
	Description    string `json:"description"`
	CoverImgURL    string `json:"cover_img_url"`
	Publish_status string `json:"publish_status"`
}

type StartPathRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type LearningPaths struct {
	PathID        string
	Title         string
	Description   string
	Instructor    string
	Students      int
	Modules       int
	Rating        float64
	CoverImgURL   string
	Objective     string
	PublishStatus string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

type LearningPath struct {
	PathID         string     `json:"path_id"`
	Title          string     `json:"title"`
	CoverImgURL    string     `json:"cover_img_url"`
	Objective      string     `json:"objective"`
	Description    string     `json:"description"`
	Rating         float64    `json:"avg_rating"`
	Publish_status string     `json:"publish_status"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	CreatorID      string     `json:"creator_id"`
	Instructor     string     `json:"instructor"`
	Modules        int        `json:"modules"`
	Students       int        `json:"student"`
}

type PathEnroll struct {
	EnrollID          string     `json:"enroll_id"`
	Enrollment_status string     `json:"enrollment_status"`
	EnrollAt          *time.Time `json:"enroll_at"`
	CompleteAt        *time.Time `json:"complete_at"`
}

type UpdateImageRequest struct {
	CoverImgURL string `json:"cover_image_url"`
}
