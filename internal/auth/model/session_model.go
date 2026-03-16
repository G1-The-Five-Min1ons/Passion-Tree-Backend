package model

import "time"

const DefaultDeactivationGracePeriod = 14

// SessionInfo represents an active user session/device
type SessionInfo struct {
	SessionID    string     `json:"session_id"`     // Same as TokenID
	DeviceInfo   string     `json:"device_info"`    // Device model/name
	IPAddress    string     `json:"ip_address"`     // Client IP
	UserAgent    string     `json:"user_agent"`     // Browser/app info
	LastUsedAt   *time.Time `json:"last_used_at"`   // Last activity
	CreatedAt    time.Time  `json:"created_at"`     // Login time
	ExpiresAt    time.Time  `json:"expires_at"`     // Token expiration
	IsCurrent    bool       `json:"is_current"`     // Is this the current session
}

// GetActiveSessionsResponse represents the response for listing active sessions
type GetActiveSessionsResponse struct {
	TotalSessions int            `json:"total_sessions"`
	Sessions      []*SessionInfo `json:"sessions"`
}
