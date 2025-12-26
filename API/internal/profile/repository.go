package user

import "database/sql"

type Repository interface {
	GetUserOverview(userID int) (*UserOverview, error)
	UpdateProfile(userID int, req UpdateProfileRequest) error
	GetSettings(userID int) (*AccountSettings, error)
	UpdatePreferences(userID int, req UpdatePreferencesRequest) error
	DeactivateAccount(userID int) error
	DeleteAccount(userID int) error
	RevokeAllSessions(userID int) error
	CreateExportJob(userID int) (string, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserOverview(userID int) (*UserOverview, error) {
	return &UserOverview{
		ID:          userID,
		Username:    "johndoe",
		DisplayName: "John Doe",
		AvatarURL:   "https://example.com/avatar.jpg",
		Status:      "active",
	}, nil
}

func (r *repository) UpdateProfile(userID int, req UpdateProfileRequest) error {
	return nil
}

func (r *repository) GetSettings(userID int) (*AccountSettings, error) {
	return &AccountSettings{
		UserID:      userID,
		Username:    "johndoe",
		Email:       "john@example.com",
		PhoneNumber: "+66812345678",
		Language:    "en",
		Region:      "TH",
	}, nil
}

func (r *repository) UpdatePreferences(userID int, req UpdatePreferencesRequest) error {
	return nil
}

func (r *repository) DeactivateAccount(userID int) error {
	return nil
}

func (r *repository) DeleteAccount(userID int) error {
	return nil
}

func (r *repository) RevokeAllSessions(userID int) error {
	return nil
}

func (r *repository) CreateExportJob(userID int) (string, error) {
	return "job_export_12345", nil
}