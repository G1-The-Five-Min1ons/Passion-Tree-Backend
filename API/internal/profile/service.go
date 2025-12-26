package user

import "errors"

type Service interface {
	GetUser(userID int) (*UserOverview, error)
	UpdateBasicProfile(userID int, req UpdateProfileRequest) error
	GetAccountSettings(userID int) (*AccountSettings, error)
	UpdateUserPreferences(userID int, req UpdatePreferencesRequest) error
	DeactivateUser(userID int) error
	DeleteUser(userID int) error
	SignOutAll(userID int) error
	ExportUserData(userID int) (string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetUser(userID int) (*UserOverview, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}
	return s.repo.GetUserOverview(userID)
}

func (s *service) UpdateBasicProfile(userID int, req UpdateProfileRequest) error {
	return s.repo.UpdateProfile(userID, req)
}

func (s *service) GetAccountSettings(userID int) (*AccountSettings, error) {
	return s.repo.GetSettings(userID)
}

func (s *service) UpdateUserPreferences(userID int, req UpdatePreferencesRequest) error {
	return s.repo.UpdatePreferences(userID, req)
}

func (s *service) DeactivateUser(userID int) error {
	return s.repo.DeactivateAccount(userID)
}

func (s *service) DeleteUser(userID int) error {
	return s.repo.DeleteAccount(userID)
}

func (s *service) SignOutAll(userID int) error {
	return s.repo.RevokeAllSessions(userID)
}

func (s *service) ExportUserData(userID int) (string, error) {
	return s.repo.CreateExportJob(userID)
}