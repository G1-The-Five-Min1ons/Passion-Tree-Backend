package model

import "time"

// TokenType constants
const (
	TokenTypeRefresh           = "refresh_token"
	TokenTypeEmailVerification = "email_verification"
	TokenTypePasswordReset     = "password_reset"
)

type Token struct {
	TokenID   string    `json:"token_id"`
	UserID    string    `json:"user_id" binding:"required"`
	Token     string    `json:"token" binding:"required"`
	TokenType string    `json:"token_type" binding:"required"`
	IsRevoked bool      `json:"is_revoked"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at" binding:"required"`
}
