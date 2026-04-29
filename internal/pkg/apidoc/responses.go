// Package apidoc holds shared response wrappers used purely for Swagger/OpenAPI
// schema generation. The runtime handlers still return fiber.Map, but the
// `@Success` / `@Failure` annotations reference these structs so the generated
// OpenAPI spec describes a stable, typed contract.
package apidoc

// SuccessResponse is the canonical envelope for successful responses.
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// MessageResponse is used when a handler returns success with no data payload.
type MessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// ErrorResponse is the canonical shape returned by Handler.handleError.
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid request body"`
}

// TokenPair represents access/refresh tokens returned from auth flows.
type TokenPair struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOi..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOi..."`
}

// TokenPairResponse is the success envelope for login / refresh / verify-email.
type TokenPairResponse struct {
	Success bool      `json:"success" example:"true"`
	Message string    `json:"message,omitempty" example:"Login successful"`
	Data    TokenPair `json:"data"`
}

// UserIDPayload is a tiny wrapper for endpoints that only return user_id.
type UserIDPayload struct {
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UserIDResponse is the envelope for endpoints that return only the user_id.
type UserIDResponse struct {
	Success bool          `json:"success" example:"true"`
	Message string        `json:"message" example:"User registered successfully"`
	Data    UserIDPayload `json:"data"`
}

// AuthURLPayload wraps an OAuth provider authorization URL.
type AuthURLPayload struct {
	AuthURL string `json:"auth_url" example:"https://accounts.google.com/o/oauth2/v2/auth?..."`
}

// AuthURLResponse is returned by the OAuth login-initiate endpoints.
type AuthURLResponse struct {
	Success bool           `json:"success" example:"true"`
	Message string         `json:"message" example:"Google OAuth URL generated"`
	Data    AuthURLPayload `json:"data"`
}

// OTPRequiredResponse is returned when login succeeds but additional OTP is required.
type OTPRequiredResponse struct {
	Success     bool   `json:"success" example:"true"`
	RequiresOTP bool   `json:"requires_otp" example:"true"`
	Message     string `json:"message" example:"OTP verification required"`
}

// ReactivationRequiredResponse is returned when login hits a deactivated account.
type ReactivationRequiredResponse struct {
	Success              bool   `json:"success" example:"true"`
	RequiresReactivation bool   `json:"requires_reactivation" example:"true"`
	Message              string `json:"message" example:"Account is deactivated"`
	GracePeriodDays      int    `json:"grace_period_days" example:"30"`
}
