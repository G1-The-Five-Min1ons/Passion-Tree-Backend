package model

import "time"

type User struct {
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	Password        string    `json:"-"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Role            UserRole  `json:"role"`
	HeartCount      int       `json:"heart_count"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Social Auth fields
	AuthProvider        string     `json:"auth_provider"`    // "local", "google", "discord"
	ProviderUserID      string     `json:"provider_user_id"` // User ID from OAuth provider
	FailedAttempts      int        `json:"failed_attempts"`
	LockedUntil         *time.Time `json:"locked_until"`
	Require2FANextLogin bool       `json:"require_2fa_next_login"` // Security flag: set to true after token theft detection
}

type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleTeacher UserRole = "teacher"
	RoleAdmin   UserRole = "admin"
	RolePending UserRole = "pending"
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
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Bio       string   `json:"bio"`
	Location  string   `json:"location"`
	AvatarURL string   `json:"avatar_url"`
	Role      UserRole `json:"role"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // รองรับทั้ง Username หรือ Email
	Password   string `json:"password"`
}

type UpdateUserRequest struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Role      UserRole `json:"role"`
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
	ProviderUserID   string
	Email            string
	FirstName        string
	LastName         string
	AvatarURL        string
	Provider         string
	ProviderUsername string
}

// NativeGoogleSignInRequest is the request body for native Google Sign-In
type NativeGoogleSignInRequest struct {
	IDToken string `json:"id_token"`
}

// NativeDiscordSignInRequest is the request body for native Discord Sign-In
type NativeDiscordSignInRequest struct {
	Code string `json:"code"`
}

// ConfirmAccountLinkRequest is the request body for confirming account linking
type ConfirmAccountLinkRequest struct {
	LinkToken string `json:"link_token"`
	Confirm   bool   `json:"confirm"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserWithProfile struct {
	User
	Profile *Profile `json:"profile,omitempty"`
}

type DashboardStats struct {
	TotalUsers       int                `json:"total_users"`
	TotalPaths       int                `json:"total_paths"`
	TotalEnrollments int                `json:"total_enrollments"`
	TotalReflections int                `json:"total_reflections"`
	RecentUsers      []*UserWithProfile `json:"recent_users"`
	UserGrowth       []int              `json:"user_growth"`
	RecentActivities []Activity         `json:"recent_activities"`
}

type Activity struct {
	Type      string    `json:"type"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
