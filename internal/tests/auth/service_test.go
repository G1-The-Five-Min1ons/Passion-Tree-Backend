package auth_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
	"passiontree/internal/auth/service"
)

// MockRepository implements repository.Repository
type mockRepository struct {
	GetUserByEmailFunc            func(ctx context.Context, email string) (*model.User, error)
	DeleteTokensByUserAndTypeFunc func(ctx context.Context, userID string, tokenType string) error
	CreateTokenFunc               func(ctx context.Context, token *model.Token) error
	RevokeTokenByValueFunc        func(ctx context.Context, tokenValue string, tokenType string) error
	GetTokenByValueFunc           func(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error)
	GetUserByIDFunc               func(ctx context.Context, id string) (*model.User, *model.Profile, error)
	ResetPasswordWithTokenFunc    func(ctx context.Context, userID string, hashedPassword string, tokenID string) error
}

// Implement RepositoryUser methods
func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}
func (m *mockRepository) CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	return "", nil
}
func (m *mockRepository) GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil, nil
}
func (m *mockRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return nil, nil
}
func (m *mockRepository) UpdateUser(ctx context.Context, id string, firstName string, lastName string, role string) error {
	return nil
}
func (m *mockRepository) UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error {
	return nil
}
func (m *mockRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	return nil
}
func (m *mockRepository) ChangePasswordAndRevokeSessions(ctx context.Context, userID string, hashedPassword string) error {
	return nil
}
func (m *mockRepository) ResetPasswordWithToken(ctx context.Context, userID string, hashedPassword string, tokenID string) error {
	if m.ResetPasswordWithTokenFunc != nil {
		return m.ResetPasswordWithTokenFunc(ctx, userID, hashedPassword, tokenID)
	}
	return nil
}
func (m *mockRepository) DeleteUser(ctx context.Context, id string) error { return nil }
func (m *mockRepository) UpdateEmailVerified(ctx context.Context, userID string, isVerified bool) error {
	return nil
}
func (m *mockRepository) VerifyEmailWithToken(ctx context.Context, userID string, tokenValue string, tokenType string) error {
	return nil
}
func (m *mockRepository) UpdateFailedLogin(ctx context.Context, userID string, lockDuration time.Duration) (int, error) {
	return 0, nil
}
func (m *mockRepository) ResetFailedLogin(ctx context.Context, userID string) error { return nil }
func (m *mockRepository) SetRequire2FANextLogin(ctx context.Context, userID string, require2FA bool) error {
	return nil
}

// Implement RepositoryToken methods
func (m *mockRepository) DeleteTokensByUserAndType(ctx context.Context, userID string, tokenType string) error {
	if m.DeleteTokensByUserAndTypeFunc != nil {
		return m.DeleteTokensByUserAndTypeFunc(ctx, userID, tokenType)
	}
	return nil
}
func (m *mockRepository) CreateToken(ctx context.Context, token *model.Token) error {
	if m.CreateTokenFunc != nil {
		return m.CreateTokenFunc(ctx, token)
	}
	return nil
}
func (m *mockRepository) RevokeTokenByValue(ctx context.Context, tokenValue string, tokenType string) error {
	if m.RevokeTokenByValueFunc != nil {
		return m.RevokeTokenByValueFunc(ctx, tokenValue, tokenType)
	}
	return nil
}
func (m *mockRepository) GetTokenByValue(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
	if m.GetTokenByValueFunc != nil {
		return m.GetTokenByValueFunc(ctx, tokenValue, tokenType)
	}
	return nil, nil
}
func (m *mockRepository) RevokeAllUserTokens(ctx context.Context, userID string, tokenType string) error {
	return nil
}
func (m *mockRepository) DeleteExpiredTokens(ctx context.Context) error { return nil }
func (m *mockRepository) MarkTokenAsRotated(ctx context.Context, tokenValue string, tokenType string) error {
	return nil
}
func (m *mockRepository) GetActiveUserSessions(ctx context.Context, userID string, tokenType string) ([]*model.Token, error) {
	return nil, nil
}
func (m *mockRepository) RevokeTokenByIDForUser(ctx context.Context, tokenID string, userID string) error {
	return nil
}
func (m *mockRepository) ReplaceVerificationToken(ctx context.Context, userID string, newToken *model.Token) error {
	return nil
}

// Implement RepositorySocial methods
func (m *mockRepository) GetUserByProvider(ctx context.Context, provider, providerUserID string) (*model.User, error) {
	return nil, nil
}
func (m *mockRepository) CreateSocialUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
	return "", nil
}
func (m *mockRepository) LinkSocialAccount(ctx context.Context, userID, provider, providerUserID string) error {
	return nil
}
func (m *mockRepository) UpdateSocialUserInfo(ctx context.Context, userID string, userInfo *model.OAuthUserInfo) error {
	return nil
}
func (m *mockRepository) UpsertSocialUserProfile(ctx context.Context, userID string, profile *model.Profile) error {
	return nil
}

// Implement GetDB
func (m *mockRepository) GetDB() repository.Database { return nil }

// MockEmailService
type mockEmailService struct {
	SendPasswordResetEmailFunc func(to, token string) error
	SendVerificationEmailFunc  func(to, token string) error
}

func (m *mockEmailService) SendPasswordResetEmail(to, token string) error {
	if m.SendPasswordResetEmailFunc != nil {
		return m.SendPasswordResetEmailFunc(to, token)
	}
	return nil
}
func (m *mockEmailService) SendVerificationEmail(to, token string) error {
	if m.SendVerificationEmailFunc != nil {
		return m.SendVerificationEmailFunc(to, token)
	}
	return nil
}
func (m *mockEmailService) SendSecurityAlertEmail(to, userID string) error { return nil }

func TestForgotPassword(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockSetup     func(*mockRepository, *mockEmailService)
		expectedError string
	}{
		{
			name:  "Success",
			email: "test@example.com",
			mockSetup: func(r *mockRepository, e *mockEmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{UserID: "user-123", Email: email}, nil
				}
				r.DeleteTokensByUserAndTypeFunc = func(ctx context.Context, userID, tokenType string) error {
					if userID != "user-123" {
						t.Errorf("Expected userID user-123, got %s", userID)
					}
					return nil
				}
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error {
					if token.UserID != "user-123" {
						t.Errorf("Expected userID user-123, got %s", token.UserID)
					}
					return nil
				}
				e.SendPasswordResetEmailFunc = func(to, token string) error {
					if to != "test@example.com" {
						t.Errorf("Expected email test@example.com, got %s", to)
					}
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:  "UserNotFound",
			email: "unknown@example.com",
			mockSetup: func(r *mockRepository, e *mockEmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, nil // User not found
				}
			},
			expectedError: "",
		},
		{
			name:  "DatabaseError",
			email: "error@example.com",
			mockSetup: func(r *mockRepository, e *mockEmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, errors.New("db disconnect")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRepository{}
			mockEmailSvc := &mockEmailService{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockEmailSvc)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			// Manually construct service with new structure
			svc := service.NewUserService(mockRepo, mockEmailSvc, nil, nil, logger)
			// Note: passing nil for config and jwtService might panic if they are used in ForgotPassword.
			// ForgotPassword uses s.config?
			// Checking auth_pwd_service.go (Step 186): It uses s.logger, s.repo, s.emailService.
			// It calls `GenerateVerificationToken` (global function?), `GetVerificationTokenExpiry`.
			// It does NOT use s.config or s.jwtService.

			// However, NewUserService returns the interface UserService.
			// We can type assert or just use the interface.
			// `svc` is UserService interface.

			err := svc.ForgotPassword(context.Background(), tt.email)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					// Check if it's an AppError and we should check .Message
					// But err.Error() returns .Message so checking err.Error() is correct for public message.
					// "internal server error"
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		newPassword   string
		mockSetup     func(*mockRepository)
		expectedError string
	}{
		{
			name:        "Success",
			code:        "valid-code",
			newPassword: "new-password-123",
			mockSetup: func(r *mockRepository) {
				// 1. GetTokenByValue
				r.GetTokenByValueFunc = func(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
					return &model.Token{
						TokenID:   "token-id",
						UserID:    "user-123",
						Token:     tokenValue,
						TokenType: model.TokenTypePasswordReset,
						ExpireAt:  time.Now().Add(1 * time.Hour), // Not expired
					}, nil
				}
				// 2. GetUserByID
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, nil, nil
				}
				// 3. ResetPasswordWithToken
				r.ResetPasswordWithTokenFunc = func(ctx context.Context, userID, hashedPassword, tokenID string) error {
					if userID != "user-123" {
						t.Errorf("Expected userID user-123, got %s", userID)
					}
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:        "ExpiredToken",
			code:        "expired-code",
			newPassword: "new-password-123",
			mockSetup: func(r *mockRepository) {
				r.GetTokenByValueFunc = func(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
					return &model.Token{
						TokenID:  "token-id",
						UserID:   "user-123",
						ExpireAt: time.Now().Add(-1 * time.Hour), // Expired
					}, nil
				}
			},
			expectedError: "reset code has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewUserService(mockRepo, nil, nil, nil, logger)

			err := svc.ResetPassword(context.Background(), tt.code, tt.newPassword)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}
