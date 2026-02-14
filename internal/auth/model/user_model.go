package model

import "time"

type User struct {
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	Password        string    `json:"-"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Role            string    `json:"role"`
	HeartCount      int       `json:"heart_count"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	FailedAttempts int        `json:"failed_attempts"`
	LockedUntil    *time.Time `json:"locked_until"`

	// Social Auth fields
	AuthProvider   string `json:"auth_provider"`    // "local", "google", "discord"
	ProviderUserID string `json:"provider_user_id"` // User ID from OAuth provider
}

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
	RoleAdmin   = "admin"
)

const (
	AuthProviderLocal   = "local"
	AuthProviderGoogle  = "google"
	AuthProviderDiscord = "discord"
)

// LinkConfirmationNeeded is returned when user needs to confirm account linking
type LinkConfirmationNeeded struct {
	LinkToken     string `json:"link_token"`
	ExistingUser  *User  `json:"existing_user"`
	ProviderEmail string `json:"provider_email"`
	ProviderName  string `json:"provider_name"`
	Message       string `json:"message"`
	NeedsConfirm  bool   `json:"needs_confirm"`
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // รองรับทั้ง Username หรือ Email
	Password   string `json:"password"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type VerifyEmailRequest struct {
	Code string `json:"code"`
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// Social Auth Models
type SocialAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	FirstName      string
	LastName       string
	AvatarURL      string
	Provider       string
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
