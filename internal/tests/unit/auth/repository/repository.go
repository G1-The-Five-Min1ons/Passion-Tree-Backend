package repository

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
)

// MockRepository implements repository.Repository
type Repository struct {
	GetUserByEmailFunc            func(ctx context.Context, email string) (*model.User, error)
	DeleteTokensByUserAndTypeFunc func(ctx context.Context, userID string, tokenType string) error
	CreateTokenFunc               func(ctx context.Context, token *model.Token) error
	RevokeTokenByValueFunc        func(ctx context.Context, tokenValue string, tokenType string) error
	GetTokenByValueFunc           func(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error)
	GetUserByIDFunc               func(ctx context.Context, id string) (*model.User, *model.Profile, error)
	ResetPasswordWithTokenFunc    func(ctx context.Context, userID string, hashedPassword string, tokenID string) error
	CreateUserFunc                func(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	UpdateUserFunc                func(ctx context.Context, id string, firstName string, lastName string, role string) error
	DeleteUserFunc                func(ctx context.Context, id string) error
	UpdateProfileFunc             func(ctx context.Context, userID string, profile *model.Profile) error
	UpdatePasswordFunc            func(ctx context.Context, userID string, hashedPassword string) error
	VerifyEmailWithTokenFunc      func(ctx context.Context, userID string, tokenValue string, tokenType string) error
	GetActiveUserSessionsFunc     func(ctx context.Context, userID string, tokenType string) ([]*model.Token, error)
	RevokeTokenByIDForUserFunc    func(ctx context.Context, tokenID string, userID string) error
	ReplaceVerificationTokenFunc  func(ctx context.Context, userID string, newToken *model.Token) error
	RevokeAllUserTokensFunc       func(ctx context.Context, userID string, tokenType string) error
	GetUserByUsernameFunc         func(ctx context.Context, username string) (*model.User, error)
	ResetFailedLoginFunc          func(ctx context.Context, userID string) error
}

// Implement RepositoryUser methods
func (m *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}
func (m *Repository) CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user, profile)
	}
	return "", nil
}
func (m *Repository) GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil, nil
}
func (m *Repository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.GetUserByUsernameFunc != nil {
		return m.GetUserByUsernameFunc(ctx, username)
	}
	return nil, nil
}
func (m *Repository) UpdateUser(ctx context.Context, id string, firstName string, lastName string, role string) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, id, firstName, lastName, role)
	}
	return nil
}
func (m *Repository) UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(ctx, userID, profile)
	}
	return nil
}
func (m *Repository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, userID, hashedPassword)
	}
	return nil
}
func (m *Repository) ChangePasswordAndRevokeSessions(ctx context.Context, userID string, hashedPassword string) error {
	return nil
}
func (m *Repository) ResetPasswordWithToken(ctx context.Context, userID string, hashedPassword string, tokenID string) error {
	if m.ResetPasswordWithTokenFunc != nil {
		return m.ResetPasswordWithTokenFunc(ctx, userID, hashedPassword, tokenID)
	}
	return nil
}
func (m *Repository) DeleteUser(ctx context.Context, id string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}
func (m *Repository) UpdateEmailVerified(ctx context.Context, userID string, isVerified bool) error {
	return nil
}
func (m *Repository) VerifyEmailWithToken(ctx context.Context, userID string, tokenValue string, tokenType string) error {
	if m.VerifyEmailWithTokenFunc != nil {
		return m.VerifyEmailWithTokenFunc(ctx, userID, tokenValue, tokenType)
	}
	return nil
}
func (m *Repository) UpdateFailedLogin(ctx context.Context, userID string, lockDuration time.Duration) (int, error) {
	return 0, nil
}
func (m *Repository) ResetFailedLogin(ctx context.Context, userID string) error {
	if m.ResetFailedLoginFunc != nil {
		return m.ResetFailedLoginFunc(ctx, userID)
	}
	return nil
}
func (m *Repository) SetRequire2FANextLogin(ctx context.Context, userID string, require2FA bool) error {
	return nil
}

// Implement RepositoryToken methods
func (m *Repository) DeleteTokensByUserAndType(ctx context.Context, userID string, tokenType string) error {
	if m.DeleteTokensByUserAndTypeFunc != nil {
		return m.DeleteTokensByUserAndTypeFunc(ctx, userID, tokenType)
	}
	return nil
}
func (m *Repository) CreateToken(ctx context.Context, token *model.Token) error {
	if m.CreateTokenFunc != nil {
		return m.CreateTokenFunc(ctx, token)
	}
	return nil
}
func (m *Repository) RevokeTokenByValue(ctx context.Context, tokenValue string, tokenType string) error {
	if m.RevokeTokenByValueFunc != nil {
		return m.RevokeTokenByValueFunc(ctx, tokenValue, tokenType)
	}
	return nil
}
func (m *Repository) GetTokenByValue(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
	if m.GetTokenByValueFunc != nil {
		return m.GetTokenByValueFunc(ctx, tokenValue, tokenType)
	}
	return nil, nil
}
func (m *Repository) RevokeAllUserTokens(ctx context.Context, userID string, tokenType string) error {
	if m.RevokeAllUserTokensFunc != nil {
		return m.RevokeAllUserTokensFunc(ctx, userID, tokenType)
	}
	return nil
}
func (m *Repository) DeleteExpiredTokens(ctx context.Context) error { return nil }
func (m *Repository) MarkTokenAsRotated(ctx context.Context, tokenValue string, tokenType string) error {
	return nil
}
func (m *Repository) GetActiveUserSessions(ctx context.Context, userID string, tokenType string) ([]*model.Token, error) {
	if m.GetActiveUserSessionsFunc != nil {
		return m.GetActiveUserSessionsFunc(ctx, userID, tokenType)
	}
	return nil, nil
}
func (m *Repository) RevokeTokenByIDForUser(ctx context.Context, tokenID string, userID string) error {
	if m.RevokeTokenByIDForUserFunc != nil {
		return m.RevokeTokenByIDForUserFunc(ctx, tokenID, userID)
	}
	return nil
}
func (m *Repository) ReplaceVerificationToken(ctx context.Context, userID string, newToken *model.Token) error {
	if m.ReplaceVerificationTokenFunc != nil {
		return m.ReplaceVerificationTokenFunc(ctx, userID, newToken)
	}
	return nil
}

// Implement RepositorySocial methods
func (m *Repository) GetUserByProvider(ctx context.Context, provider, providerUserID string) (*model.User, error) {
	return nil, nil
}
func (m *Repository) CreateSocialUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	return "", nil
}
func (m *Repository) LinkSocialAccount(ctx context.Context, userID, provider, providerUserID string) error {
	return nil
}
func (m *Repository) UpdateSocialUserInfo(ctx context.Context, userID string, userInfo *model.OAuthUserInfo) error {
	return nil
}
func (m *Repository) UpsertSocialUserProfile(ctx context.Context, userID string, profile *model.Profile) error {
	return nil
}

// Implement GetDB
func (m *Repository) GetDB() repository.Database { return nil }
