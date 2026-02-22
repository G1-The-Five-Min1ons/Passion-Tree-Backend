package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/service"
	"passiontree/internal/config"
	"passiontree/internal/pkg/jwt"
	repository_test "passiontree/internal/tests/unit/auth/repository"
)

func TestVerifyEmail(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:          "testsecret",
		JWTAccessTTL:       "1",
		JWTRefreshTTL:      "168",
		JWTRefreshAbsolute: "720",
	}
	jwtSvc := jwt.NewService(cfg)

	tests := []struct {
		name          string
		vToken        string
		deviceInfo    string
		ip            string
		ua            string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:       "Success",
			vToken:     "valid-verify-token",
			deviceInfo: "TestDevice",
			ip:         "127.0.0.1",
			ua:         "TestAgent",
			setup: func(r *repository_test.Repository) {
				r.GetTokenByValueFunc = func(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
					return &model.Token{
						TokenID:   "token-1",
						UserID:    "user-1",
						Token:     tokenValue,
						TokenType: model.TokenTypeEmailVerification,
						ExpireAt:  time.Now().Add(1 * time.Hour),
					}, nil
				}
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id, Role: "user", FirstName: "Test"}, &model.Profile{}, nil
				}
				r.VerifyEmailWithTokenFunc = func(ctx context.Context, userID, tokenValue, tokenType string) error {
					return nil
				}
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:   "ExpiredToken",
			vToken: "expired-token",
			setup: func(r *repository_test.Repository) {
				r.GetTokenByValueFunc = func(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
					return &model.Token{
						TokenID:   "token-2",
						UserID:    "user-2",
						TokenType: model.TokenTypeEmailVerification,
						ExpireAt:  time.Now().Add(-1 * time.Hour), // Past Time
					}, nil
				}
			},
			expectedError: "verification code has expired",
		},
		{
			name:   "TokenNotFound",
			vToken: "invalid-token",
			setup: func(r *repository_test.Repository) {
				r.GetTokenByValueFunc = func(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
					return nil, nil // Not found
				}
			},
			expectedError: "invalid or expired verification code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestVerifyEmail case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewUserService(mockRepo, nil, cfg, jwtSvc, logger)

			accToken, refToken, err := svc.VerifyEmail(context.Background(), tt.vToken, tt.deviceInfo, tt.ip, tt.ua)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if accToken == "" || refToken == "" {
					t.Errorf("Expected tokens to be generated, got empty strings")
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

func TestResendVerificationEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		setup         func(*repository_test.Repository, *mockEmailService)
		expectedError string
	}{
		{
			name:  "Success_NotVerified",
			email: "unverified@example.com",
			setup: func(r *repository_test.Repository, e *mockEmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{UserID: "user-1", Email: email, IsEmailVerified: false}, nil
				}
				r.ReplaceVerificationTokenFunc = func(ctx context.Context, userID string, newToken *model.Token) error {
					return nil
				}
				e.SendVerificationEmailFunc = func(to, token string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:  "AlreadyVerified",
			email: "verified@example.com",
			setup: func(r *repository_test.Repository, e *mockEmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{UserID: "user-2", Email: email, IsEmailVerified: true}, nil
				}
			},
			expectedError: "email already verified",
		},
		{
			name:  "UserNotFound",
			email: "notfound@example.com",
			setup: func(r *repository_test.Repository, e *mockEmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, nil // Returns nil for not found
				}
			},
			expectedError: "user with this email does not exist", // Service usually swallows this to prevent enum attacks
		},
		{
			name:  "DatabaseError",
			email: "error@example.com",
			setup: func(r *repository_test.Repository, e *mockEmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, errors.New("db down")
				}
			},
			expectedError: "user with this email does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestResendVerificationEmail case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			mockEmailSvc := &mockEmailService{}

			if tt.setup != nil {
				tt.setup(mockRepo, mockEmailSvc)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			cfg := &config.Config{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, cfg, nil, logger)

			err := svc.ResendVerificationEmail(context.Background(), tt.email)

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
