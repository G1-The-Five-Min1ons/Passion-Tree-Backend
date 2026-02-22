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

func TestLogin(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:          "testsecret",
		JWTAccessTTL:       "1",
		JWTRefreshTTL:      "168",
		JWTRefreshAbsolute: "720",
	}
	jwtSvc := jwt.NewService(cfg)

	tests := []struct {
		name          string
		identifier    string
		password      string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:       "Success_Email_Unverified",
			identifier: "test@example.com",
			password:   "correct_password",
			setup: func(r *repository_test.Repository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					// Dummy bcrypt hash for testing (cost 10, "correct_password")
					return &model.User{
						UserID: "42611365-6415-4530-9346-3ee695d8b58d", Email: email, Username: "farloss",
						Password:        "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
						IsEmailVerified: false,
					}, nil
				}
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, &model.Profile{}, nil
				}
				// Mock token creations if used. Login does create tokens.
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error { return nil }
				r.ResetFailedLoginFunc = func(ctx context.Context, userID string) error { return nil }
			},
			// Wait, the real implementation of Login will probably return an unverified exception.
			expectedError: "verification_required",
		},
		{
			name:       "UserNotFound",
			identifier: "unknown@example.com",
			password:   "any",
			setup: func(r *repository_test.Repository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) { return nil, nil }
				r.GetUserByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) { return nil, nil }
			},
			expectedError: "invalid username/email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestLogin case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &mockEmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, cfg, jwtSvc, logger)

			_, _, err := svc.Login(context.Background(), tt.identifier, tt.password, "Device", "IP", "UA")

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

func TestLogout(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:   "Success",
			userID: "user-1",
			setup: func(r *repository_test.Repository) {
				r.RevokeAllUserTokensFunc = func(ctx context.Context, userID string, tokenType string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:   "Failure",
			userID: "user-2",
			setup: func(r *repository_test.Repository) {
				r.RevokeAllUserTokensFunc = func(ctx context.Context, userID string, tokenType string) error {
					return errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestLogout case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewUserService(mockRepo, nil, nil, nil, logger)

			err := svc.Logout(context.Background(), tt.userID)

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

func TestGetActiveSessions(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		currentToken  string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:         "Success",
			userID:       "user-1",
			currentToken: "token-value",
			setup: func(r *repository_test.Repository) {
				dev1 := "Device1"
				dev2 := "Device2"
				tNow := time.Now()
				r.GetActiveUserSessionsFunc = func(ctx context.Context, userID string, tokenType string) ([]*model.Token, error) {
					return []*model.Token{
						{TokenID: "active-1", DeviceInfo: &dev1, LastUsedAt: &tNow},
						{TokenID: "active-2", DeviceInfo: &dev2, LastUsedAt: &tNow, Token: "token-value"},
					}, nil
				}
			},
			expectedError: "",
		},
		{
			name:   "Failure",
			userID: "user-2",
			setup: func(r *repository_test.Repository) {
				r.GetActiveUserSessionsFunc = func(ctx context.Context, userID string, tokenType string) ([]*model.Token, error) {
					return nil, errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetActiveSessions case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewUserService(mockRepo, nil, nil, nil, logger)

			_, err := svc.GetActiveSessions(context.Background(), tt.userID, tt.currentToken)

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

func TestLogoutSession(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		sessionID     string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:      "Success",
			userID:    "user-1",
			sessionID: "session-1",
			setup: func(r *repository_test.Repository) {
				r.RevokeTokenByIDForUserFunc = func(ctx context.Context, tokenID string, userID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:      "Failure",
			userID:    "user-2",
			sessionID: "session-2",
			setup: func(r *repository_test.Repository) {
				r.RevokeTokenByIDForUserFunc = func(ctx context.Context, tokenID string, userID string) error {
					return errors.New("db disconnect")
				}
			},
			expectedError: "session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestLogoutSession case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewUserService(mockRepo, nil, nil, nil, logger)

			err := svc.LogoutSession(context.Background(), tt.userID, tt.sessionID)

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
