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
	UpdateFailedLogin(ctx context.Context, userID string, lockDuration time.Duration) (int, error)
	ResetFailedLogin(ctx context.Context, userID string) error

	GetDB() *sql.DB
}

type TokenRepository interface {
	CreateToken(token *model.Token) error
	GetTokenByValue(tokenValue string, tokenType string) (*model.Token, error)
	RevokeTokenByValue(tokenValue string, tokenType string) error
	RevokeAllUserTokens(userID string, tokenType string) error
	DeleteExpiredTokens() error
	DeleteTokensByUserAndType(userID string, tokenType string) error
	
	// Token Rotation Methods
	MarkTokenAsRotated(tokenValue string, tokenType string) error
	
	// Multi-device Session Management
	GetActiveUserSessions(userID string, tokenType string) ([]*model.Token, error)
	RevokeTokenByIDForUser(tokenID string, userID string) error
}

type userRepositoryImpl struct {
	db *sql.DB
}

type tokenRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(ds connection.Database) UserRepository {
	return &userRepositoryImpl{
		db: ds.GetDB(),
	}
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepositoryImpl{
		db: db,
	}
}

// GetDB returns the database connection for direct queries when needed
func (r *userRepositoryImpl) GetDB() *sql.DB {
	return r.db
}
