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

type LearningPath struct {
	PathID         string     `json:"path_id"`
	Title          string     `json:"title"`
	CoverImgURL    string     `json:"cover_img_url"`
	Objective      string     `json:"objective"`
	Description    string     `json:"description"`
	AvgRating      float64    `json:"avg_rating"`
	Publish_status string     `json:"publish_status"`
	CreatedAt      *time.Time `json:"create_at"`
	UpdatedAt      *time.Time `json:"update_at"`
	CreatorID      string     `json:"creator_id"`
}

type PathEnroll struct {
	EnrollID          string     `json:"enroll_id"`
	Enrollment_status string     `json:"enrollment_status"`
	EnrollAt          *time.Time `json:"enroll_at"`
	CompleteAt        *time.Time `json:"complete_at"`
}
