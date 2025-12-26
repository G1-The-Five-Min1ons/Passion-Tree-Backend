package user

type UserOverview struct {
	ID          int    `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Status      string `json:"status"`
}

type AccountSettings struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Language    string `json:"language"`
	Region      string `json:"region"`
}

type UserPreferences struct {
	TwoFactorEnabled bool   `json:"two_factor_enabled"`
	AutoSave         bool   `json:"auto_save"`
	TimeZone         string `json:"time_zone"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
}

type UpdatePreferencesRequest struct {
	TwoFactorEnabled bool   `json:"two_factor_enabled"`
	AutoSave         bool   `json:"auto_save"`
	TimeZone         string `json:"time_zone"`
}

type SignOutAllRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

type ExportRequest struct {
	UserID int `json:"user_id" binding:"required"`
}