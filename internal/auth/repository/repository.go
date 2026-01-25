package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/auth/model"
	"passiontree/internal/database"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
}

type userRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(ds database.Database) UserRepository {
	return &userRepositoryImpl{
		db: ds.GetDB(),
	}
}
