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
		setup         func(*repository_test.Repository, *EmailService)
		expectedError string
	}{
		{
			name:          "MissingIdentifier",
			identifier:    "",
			password:      "any",
			expectedError: "identifier and password are required",
		},
		{
			name:          "UserNotFound_ByUsername",
			identifier:    "invalid-email",
			password:      "any",
			expectedError: "invalid username/email or password",
		},
		{
			name:       "IncorrectPassword",
			identifier: "test@example.com",
			password:   "wrong_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{
						UserID: "42611365-6415-4530-9346-3ee695d8b58d", Email: email, Username: "farloss",
						Password: "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
					}, nil
				}
			},
			expectedError: "invalid username/email or password",
		},
		{
			name:       "AccountLocked",
			identifier: "locked@example.com",
			password:   "any",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					lockTime := time.Now().Add(15 * time.Minute)
					return &model.User{
						UserID: "locked-user-id", Email: email, Username: "lockeduser",
						LockedUntil: &lockTime,
					}, nil
				}
			},
			expectedError: "account locked",
		},
		{
			name:       "2FAMandatory",
			identifier: "2fa@example.com",
			password:   "correct_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{
						UserID: "u-2fa", Email: email, Username: "2fauser",
						Password:            "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
						Require2FANextLogin: true, // ตัวจุดชนวน 2FA
					}, nil
				}
				r.DeleteTokensByUserAndTypeFunc = func(ctx context.Context, uid, tType string) error { return nil }
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error { return nil }
			},
			expectedError: "security verification required",
		},
		{
			name:       "SuccessEmailUnverified",
			identifier: "test@example.com",
			password:   "correct_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
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
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) { return nil, nil }
				r.GetUserByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) { return nil, nil }
			},
			expectedError: "invalid username/email or password",
		},
		{
			name:       "CreateTokenError",
			identifier: "test@example.com",
			password:   "correct_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{
						UserID: "42611365-6415-4530-9346-3ee695d8b58d", Email: email, Username: "farloss",
						Password:        "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
						IsEmailVerified: true,
					}, nil
				}
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error {
					return errors.New("token creation error")
				}
			},
			expectedError: "internal server error",
		},
		{
			name:       "2FASessionCreationSuccess",
			identifier: "2fa@example.com",
			password:   "correct_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{
						UserID:              "u-2fa-123",
						Email:               email,
						Password:            "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
						Require2FANextLogin: true, //
					}, nil
				}

				r.DeleteTokensByUserAndTypeFunc = func(ctx context.Context, uid string, tType string) error {
					return nil
				}

				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error {
					if token.TokenType != model.TokenType2FA {
						return errors.New("expected 2FA token type")
					}
					return nil
				}
			},
			expectedError: "security verification required",
		},
		{
			name:       "2FASessionCreationFailure",
			identifier: "2fa@example.com",
			password:   "correct_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{
						UserID:              "u-2fa-123",
						Email:               email,
						Password:            "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
						Require2FANextLogin: true, //
					}, nil
				}

				r.DeleteTokensByUserAndTypeFunc = func(ctx context.Context, uid string, tType string) error {
					return nil
				}

				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error {
					return errors.New("token creation failed")
				}
			},
			expectedError: "internal server error",
		},
		{
			name:       "LoginSuccess_ByUsername",
			identifier: "farloss", // ไม่มี @ เพื่อให้ระบบเรียก GetUserByUsername
			password:   "correct_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return &model.User{
						UserID: "u-1", Email: "test@example.com", Username: username,
						Password: "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
					}, nil
				}
				r.DeleteTokensByUserAndTypeFunc = func(ctx context.Context, uid, tType string) error { return nil }
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error { return nil }
			},
			expectedError: "verification_required", // ติดที่ด่าน OTP
		},
		{
			name:       "EmailServiceFailure",
			identifier: "test@example.com",
			password:   "correct_password",
			setup: func(r *repository_test.Repository, e *EmailService) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{
						UserID: "u-1", Email: email,
						Password: "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa",
					}, nil
				}
				r.DeleteTokensByUserAndTypeFunc = func(ctx context.Context, uid, tType string) error { return nil }
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error { return nil }
				e.SendVerificationEmailFunc = func(ctx context.Context, email, token string) error {
					return errors.New("smtp server down")
				}
			},
			expectedError: "verification_required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestLogin case: %s\033[0m", tt.name)
			Repo := &repository_test.Repository{}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			EmailSvc := &EmailService{}
			svc := service.NewUserService(Repo, EmailSvc, cfg, jwtSvc, logger)

			if tt.setup != nil {
				tt.setup(Repo, EmailSvc)
			}

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
		{
			name:   "RevokeTokensFailure",
			userID: "user-1",
			setup: func(r *repository_test.Repository) {
				r.RevokeAllUserTokensFunc = func(ctx context.Context, userID string, tokenType string) error {
					return errors.New("database connection lost")
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
			name:      "SessionNotFound",
			userID:    "user-1",
			sessionID: "nonexistent-session",
			setup: func(r *repository_test.Repository) {
				r.RevokeTokenByIDForUserFunc = func(ctx context.Context, tokenID string, userID string) error {
					return errors.New("session not found")
				}
			},
			expectedError: "session not found",
		},
		{
			name:      "LogoutSuccess",
			userID:    "user-1",
			sessionID: "session-1",
			setup: func(r *repository_test.Repository) {
				r.RevokeTokenByIDForUserFunc = func(ctx context.Context, tokenID string, userID string) error {
					return nil
				}
			},
			expectedError: "",
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

func TestRefreshAccessToken(t *testing.T) {
	// 1. Setup พื้นฐานที่จำเป็น
	cfg := &config.Config{
		JWTSecret:     "testsecret",
		JWTAccessTTL:  "1",
		JWTRefreshTTL: "168",
	}
	jwtSvc := jwt.NewService(cfg)

	// 2. กำหนดตารางการทดสอบ (Table-driven tests)
	tests := []struct {
		name          string
		refreshToken  string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:         "Success_Rotation",
			refreshToken: "valid-token",
			setup: func(r *repository_test.Repository) {
				// Mock ให้เจอ Token ปกติที่ยังไม่เคยถูกใช้งาน
				r.GetTokenByValueFunc = func(ctx context.Context, val string, tType string) (*model.Token, error) {
					return &model.Token{TokenID: "t-1", UserID: "u-1", IsRotated: false, IsRevoked: false}, nil
				}
				// Mock ให้เจอ User
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id, Username: "testuser"}, &model.Profile{}, nil
				}
				r.MarkTokenAsRotatedFunc = func(ctx context.Context, val string, tType string) error { return nil }
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error { return nil }
			},
			expectedError: "",
		},
		{
			name:         "TheftDetection_TokenReused",
			refreshToken: "stolen-token",
			setup: func(r *repository_test.Repository) {
				// ตรวจเจอว่า Token นี้ถูกหมุนเวียนไปแล้ว (IsRotated: true) -> สัญญาณโดนขโมย
				r.GetTokenByValueFunc = func(ctx context.Context, val string, tType string) (*model.Token, error) {
					return &model.Token{UserID: "u-hack", IsRotated: true}, nil
				}
				// ระบบต้องล้าง Session ทั้งหมด (handleTokenTheft)
				r.RevokeAllUserTokensFunc = func(ctx context.Context, uid, tType string) error { return nil }
				r.SetRequire2FANextLoginFunc = func(ctx context.Context, uid string, req bool) error { return nil }
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id, Email: "victim@test.com"}, nil, nil
				}
			},
			expectedError: "security violation detected - all sessions terminated", //
		},
		{
			name:         "AbsoluteExpiration_Exceeded",
			refreshToken: "old-token",
			setup: func(r *repository_test.Repository) {
				expiredMax := time.Now().Add(-1 * time.Hour) // เวลาในอดีต
				r.GetTokenByValueFunc = func(ctx context.Context, val string, tType string) (*model.Token, error) {
					return &model.Token{UserID: "u-1", MaxExpiresAt: &expiredMax}, nil
				}
				r.RevokeTokenByValueFunc = func(ctx context.Context, val string, tType string) error { return nil }
			},
			expectedError: "session expired", //
		},
	}

	// 3. ลูปการรันเทส (Execution Loop)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &EmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, cfg, jwtSvc, logger)

			testUser := &model.User{UserID: "u-1"}
			token, _ := jwtSvc.GenerateRefreshToken(testUser)
			if tt.refreshToken != "" && tt.refreshToken != "invalid" {
				tt.refreshToken = token
			}

			_, _, err := svc.RefreshAccessToken(context.Background(), tt.refreshToken, "Dev", "IP", "UA")

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

func TestValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:    "testsecret",
		JWTAccessTTL: "1",
	}
	jwtSvc := jwt.NewService(cfg)

	tests := []struct {
		name          string
		token         string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:          "EmptyToken",
			token:         "",
			expectedError: "token is required",
		},
		{
			name:          "InvalidToken",
			token:         "invalid.jwt.token",
			expectedError: "invalid token",
		},
		{
			name:  "UserNotFound",
			token: "valid-token",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return nil, nil, nil
				}
			},
			expectedError: "user not found",
		},
		{
			name:  "DatabaseError",
			token: "valid-token",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return nil, nil, errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
		{
			name:  "Success",
			token: "valid-token",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, nil, nil
				}
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestValidateToken case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewUserService(mockRepo, nil, cfg, jwtSvc, logger)

			testUser := &model.User{UserID: "u-1"}
			if tt.token == "valid-token" {
				tt.token, _ = jwtSvc.GenerateAccessToken(testUser)
			}

			_, err := svc.ValidateToken(context.Background(), tt.token)

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
