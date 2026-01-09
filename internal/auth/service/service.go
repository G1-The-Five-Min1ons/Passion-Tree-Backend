package service

import (
	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
)

type UserService interface {
	CreateUser(user *model.User, profile *model.Profile) (string, error)
	GetUserByID(id string) (*model.User, *model.Profile, error)
	GetUserByEmail(email string) (*model.User, error)
	UpdateUser(id string, user *model.User) error
	UpdateProfile(userID string, profile *model.Profile) error
	DeleteUser(id string) error
	Login(email string, password string) (string, error)
	ValidateToken(token string) (*model.User, error)
}

type userServiceImpl struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
	}
}
