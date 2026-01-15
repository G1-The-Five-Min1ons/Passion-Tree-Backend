package service

import (
	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
)

type UserService interface {
	CreateUser(user *model.User, profile *model.Profile) (string, error)
	GetUserByID(id string) (*model.User, *model.Profile, error)
	GetUserByEmail(email string) (*model.User, error)
	UpdateUser(id string, firstName string, lastName string) error
	UpdateProfile(userID string, profile *model.Profile) error
	DeleteUser(id string, password string) error
	Login(identifier string, password string) (string, error)
	ValidateToken(token string) (*model.User, error)
	VerifyEmail(token string) error
	ResendVerificationEmail(email string) error
}

type userServiceImpl struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	emailService EmailService
}

func NewUserService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository) UserService {
	return &userServiceImpl{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

// SetEmailService sets the email service (used for dependency injection)
func (s *userServiceImpl) SetEmailService(emailService EmailService) {
	s.emailService = emailService
}
