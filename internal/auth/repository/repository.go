package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/auth/model"
	"passiontree/internal/connection"
	"time"
)

// Repository combines all auth-related repository interfaces
type Repository interface {
	UserRepository
	TokenRepository
	SocialAuthRepository
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, firstName string, lastName string) error
	UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
	ChangePasswordAndRevokeSessions(ctx context.Context, userID string, hashedPassword string) error
	ResetPasswordWithToken(ctx context.Context, userID string, hashedPassword string, tokenID string) error
	DeleteUser(ctx context.Context, id string) error
	UpdateEmailVerified(ctx context.Context, userID string, isVerified bool) error
	VerifyEmailAndRevokeToken(ctx context.Context, userID string, tokenID string) error
	UpdateFailedLogin(ctx context.Context, userID string, lockDuration time.Duration) (int, error)
	ResetFailedLogin(ctx context.Context, userID string) error

	GetDB() *sql.DB
}

type TokenRepository interface {
	CreateToken(ctx context.Context, token *model.Token) error
	GetTokenByValue(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error)
	RevokeToken(ctx context.Context, tokenID string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteTokensByUserAndType(ctx context.Context, userID string, tokenType string) error
}

type SocialAuthRepository interface {
	GetUserByProvider(ctx context.Context, provider, providerUserID string) (*model.User, error)
	CreateSocialUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	LinkSocialAccount(ctx context.Context, userID, provider, providerUserID string) error
	UpdateSocialUserInfo(ctx context.Context, userID string, userInfo *model.OAuthUserInfo) error
	UpsertSocialUserProfile(ctx context.Context, userID string, profile *model.Profile) error
}

// repositoryImpl implements all repository interfaces
type repositoryImpl struct {
	*userRepositoryImpl
	*tokenRepositoryImpl
	*socialAuthRepositoryImpl
}

// NewRepository creates a new unified repository instance
func NewRepository(ds connection.Database) Repository {
	db := ds.GetDB()
	return &repositoryImpl{
		userRepositoryImpl:       &userRepositoryImpl{db: db},
		tokenRepositoryImpl:      &tokenRepositoryImpl{db: db},
		socialAuthRepositoryImpl: &socialAuthRepositoryImpl{db: db},
	}
}

type userRepositoryImpl struct {
	db *sql.DB
}

// GetDB returns the database connection for direct queries when needed
func (r *userRepositoryImpl) GetDB() *sql.DB {
	return r.db
}

type tokenRepositoryImpl struct {
	db *sql.DB
}

type socialAuthRepositoryImpl struct {
	db *sql.DB
}
