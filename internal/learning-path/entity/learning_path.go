package entity

import "time"

// LearningPath represents a learning path in the system
type LearningPath struct {
	LearningPathID string    `json:"learning_path_id" db:"LearningPathID"`
	UserID         string    `json:"user_id" db:"UserID"`
	Title          string    `json:"title" db:"Title"`
	Description    *string   `json:"description" db:"Description"` // Nullable
	CreatedAt      time.Time `json:"created_at" db:"CreatedAt"`
	UpdatedAt      time.Time `json:"updated_at" db:"UpdatedAt"`
}

// CreateLearningPathRequest represents the request to create a new learning path
type CreateLearningPathRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

// UpdateLearningPathRequest represents the request to update a learning path
type UpdateLearningPathRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

// LearningPathResponse represents the learning path response
type LearningPathResponse struct {
	LearningPathID string    `json:"learning_path_id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ToResponse converts LearningPath to LearningPathResponse
func (lp *LearningPath) ToResponse() *LearningPathResponse {
	return &LearningPathResponse{
		LearningPathID: lp.LearningPathID,
		UserID:         lp.UserID,
		Title:          lp.Title,
		Description:    lp.Description,
		CreatedAt:      lp.CreatedAt,
		UpdatedAt:      lp.UpdatedAt,
	}
}