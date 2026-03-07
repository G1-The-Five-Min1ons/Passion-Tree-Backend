package model

import "time"

type ApplyTeacherRequest struct {
	PhoneNumber     string `json:"phone_number"`
	Reason          string `json:"reason"`
	TeachingHistory string `json:"teaching_history"`
}

type TeacherVerificationStatus struct {
	PhoneNumber       string `json:"phone_number"`
	HasPhoneNumber    bool   `json:"has_phone_number"`
	HasApplied        bool   `json:"has_applied"`
	ApplicationStatus string `json:"application_status"`
	IsVerified        bool   `json:"is_verified"`
}

type TeacherVerificationRequest struct {
	RequestID       string     `json:"request_id"`
	UserID          string     `json:"user_id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	PhoneNumber     string     `json:"phone_number"`
	Reason          string     `json:"reason"`
	TeachingHistory string     `json:"teaching_history"`
	Status          string     `json:"status"`
	ReviewedBy      string     `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ReviewTeacherApplicationRequest struct {
	Status string `json:"status"`
}
