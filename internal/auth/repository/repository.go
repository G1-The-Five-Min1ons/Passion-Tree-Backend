package repository

import (
	"database/sql"
	"passiontree/internal/auth/model"
	"passiontree/internal/database"
)

type UserRepository interface {
	CreateUser(user *model.User, profile *model.Profile) (string, error)
	GetUserByID(id string) (*model.User, *model.Profile, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	UpdateUser(id string, firstName string, lastName string) error
	UpdateProfile(userID string, profile *model.Profile) error
	DeleteUser(id string) error
}

type userRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(ds database.Database) UserRepository {
	return &userRepositoryImpl{
		db: ds.GetDB(),
	}
}
