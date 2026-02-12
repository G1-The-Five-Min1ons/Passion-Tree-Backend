package repository

import (
	"time"
	"context"
	"database/sql"
	"passiontree/internal/auth/model"
	"passiontree/internal/connection"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, firstName string, lastName string) error
	UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error
	DeleteUser(ctx context.Context, id string) error
	UpdateEmailVerified(ctx context.Context, userID string, isVerified bool) error
	UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error
	ResetFailedLogin(ctx context.Context, userID string) error

	GetDB() *sql.DB
}

type userRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(ds connection.Database) UserRepository {
	return &userRepositoryImpl{
		db: ds.GetDB(),
	}
}

// GetDB returns the database connection for direct queries when needed
func (r *userRepositoryImpl) GetDB() *sql.DB {
	return r.db
}
