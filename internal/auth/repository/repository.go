package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/auth/model"
	"passiontree/internal/connection"
	"time"
)

type Repository interface {
	// User Repository Methods
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

	// Token Repository Methods
	CreateToken(ctx context.Context, token *model.Token) error
	GetTokenByValue(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error)
	RevokeTokenByValue(ctx context.Context, tokenValue string, tokenType string) error
	RevokeAllUserTokens(ctx context.Context, userID string, tokenType string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteTokensByUserAndType(ctx context.Context, userID string, tokenType string) error
	MarkTokenAsRotated(ctx context.Context, tokenValue string, tokenType string) error
	GetActiveUserSessions(ctx context.Context, userID string, tokenType string) ([]*model.Token, error)
	RevokeTokenByIDForUser(ctx context.Context, tokenID string, userID string) error

	// Password Management Methods
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
	ResetPasswordWithToken(ctx context.Context, userID string, hashedPassword string, tokenID string) error

	// Email Verification with Transaction
	VerifyEmailWithToken(ctx context.Context, userID string, tokenValue string, tokenType string) error
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
