package entity

import "time"

// User represents a user in the system
type User struct {
	UserID       string    `json:"user_id" db:"UserID"`
	Email        string    `json:"email" db:"Email"`
	PasswordHash string    `json:"-" db:"PasswordHash"` // Never send password in JSON
	CreatedAt    time.Time `json:"created_at" db:"CreatedAt"`
	UpdatedAt    time.Time `json:"updated_at" db:"UpdatedAt"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse represents the user response (without password)
type UserResponse struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		UserID:    u.UserID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
