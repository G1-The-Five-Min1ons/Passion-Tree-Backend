package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/auth/model"
	"passiontree/internal/connection"
	"time"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Database interface {
	DBTX
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

type RepositoryUser interface {
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
	VerifyEmailWithToken(ctx context.Context, userID string, tokenValue string, tokenType string) error 
	UpdateFailedLogin(ctx context.Context, userID string, lockDuration time.Duration) (int, error)
	ResetFailedLogin(ctx context.Context, userID string) error
	SetRequire2FANextLogin(ctx context.Context, userID string, require2FA bool) error 
}

type RepositoryToken interface {
	CreateToken(ctx context.Context, token *model.Token) error
	GetTokenByValue(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error)
	RevokeTokenByValue(ctx context.Context, tokenValue string, tokenType string) error 
	RevokeAllUserTokens(ctx context.Context, userID string, tokenType string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteTokensByUserAndType(ctx context.Context, userID string, tokenType string) error
	MarkTokenAsRotated(ctx context.Context, tokenValue string, tokenType string) error        
	GetActiveUserSessions(ctx context.Context, userID string, tokenType string) ([]*model.Token, error) 
	RevokeTokenByIDForUser(ctx context.Context, tokenID string, userID string) error
	ReplaceVerificationToken(ctx context.Context, userID string, newToken *model.Token) error  
}

type RepositorySocial interface {
	GetUserByProvider(ctx context.Context, provider, providerUserID string) (*model.User, error)
	CreateSocialUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	LinkSocialAccount(ctx context.Context, userID, provider, providerUserID string) error
	UpdateSocialUserInfo(ctx context.Context, userID string, userInfo *model.OAuthUserInfo) error
	UpsertSocialUserProfile(ctx context.Context, userID string, profile *model.Profile) error
}

type Repository interface {
	RepositoryUser
	RepositoryToken
	RepositorySocial
	GetDB() Database
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB().(Database),
	}
}

func (r *repositoryImpl) GetDB() Database {
	return r.db
}