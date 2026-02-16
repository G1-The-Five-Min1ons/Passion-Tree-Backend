package model

import "time"

// TokenType constants
const (
	TokenTypeRefresh           = "refresh_token"
	TokenTypeEmailVerification = "email_verification"
	TokenTypePasswordReset     = "password_reset"
)

type Token struct {
	TokenID       string     `json:"token_id"`
	UserID        string     `json:"user_id" binding:"required"`
	Token         string     `json:"token" binding:"required"`
	TokenType     string     `json:"token_type" binding:"required"`
	IsRevoked     bool       `json:"is_revoked"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpireAt      time.Time  `json:"expire_at" binding:"required"`
	
	// Token Rotation and Session Tracking Fields
	DeviceInfo    *string    `json:"device_info,omitempty"`     // Device model/name
	IPAddress     *string    `json:"ip_address,omitempty"`      // Client IP address
	UserAgent     *string    `json:"user_agent,omitempty"`      // Browser/app user agent
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`    // Last time token was used
	MaxExpiresAt  *time.Time `json:"max_expires_at,omitempty"`  // Absolute expiration (for sliding window)
	ParentTokenID *string    `json:"parent_token_id,omitempty"` // ID of the token that was rotated to create this one
	IsRotated     bool       `json:"is_rotated"`                // Whether this token has been rotated (replaced)
}
