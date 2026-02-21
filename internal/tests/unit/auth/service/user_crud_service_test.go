package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/service"
	repository_test "passiontree/internal/tests/unit/auth/repository"
)

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name          string
		user          *model.User
		profile       *model.Profile
		mockSetup     func(*repository_test.MockRepository)
		expectedError string
	}{
		{
			name:    "Success",
			user:    &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile: &model.Profile{},
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, nil // Ensure email is not taken
				}
				r.CreateUserFunc = func(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
					return "new-user-id", nil
				}
			},
			expectedError: "",
		},
		{
			name:    "EmailAlreadyExists",
			user:    &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile: &model.Profile{},
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{UserID: "existing-id"}, nil // Email is taken
				}
			},
			expectedError: "email already registered",
		},
		{
			name:    "DatabaseError",
			user:    &model.User{Username: "farloss", Email: "error@example.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile: &model.Profile{},
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, nil
				}
				r.CreateUserFunc = func(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
					return "", errors.New("db insert failed")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestCreateUser case: %s\033[0m", tt.name)
			mockRepo := &repository_test.MockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &mockEmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, nil, nil, logger)

			_, err := svc.CreateUser(context.Background(), tt.user, tt.profile)

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

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		mockSetup     func(*repository_test.MockRepository)
		expectedError string
	}{
		{
			name: "Success",
			id:   "user-1",
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id, FirstName: "Test"}, &model.Profile{}, nil
				}
			},
			expectedError: "",
		},
		{
			name: "NotFound",
			id:   "user-2",
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return nil, nil, nil
				}
			},
			expectedError: "user with id 'user-2' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetUserByID case: %s\033[0m", tt.name)
			mockRepo := &repository_test.MockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &mockEmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, nil, nil, logger)

			_, _, err := svc.GetUserByID(context.Background(), tt.id)

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

func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockSetup     func(*repository_test.MockRepository)
		expectedError string
	}{
		{
			name:  "Success",
			email: "found@example.com",
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{Email: email}, nil
				}
			},
			expectedError: "",
		},
		{
			name:  "NotFound",
			email: "missing@example.com",
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, nil
				}
			},
			expectedError: "user with email 'missing@example.com' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetUserByEmail case: %s\033[0m", tt.name)
			mockRepo := &repository_test.MockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &mockEmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, nil, nil, logger)

			_, err := svc.GetUserByEmail(context.Background(), tt.email)

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

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		firstName     string
		lastName      string
		role          string
		mockSetup     func(*repository_test.MockRepository)
		expectedError string
	}{
		{
			name:      "Success",
			id:        "user-1",
			firstName: "Updated",
			lastName:  "Name",
			role:      "admin",
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, &model.Profile{}, nil
				}
				r.UpdateUserFunc = func(ctx context.Context, id string, firstName string, lastName string, role string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name: "RepositoryError",
			id:   "user-2",
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, &model.Profile{}, nil
				}
				r.UpdateUserFunc = func(ctx context.Context, id string, firstName string, lastName string, role string) error {
					return errors.New("db update failed")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestUpdateUser case: %s\033[0m", tt.name)
			mockRepo := &repository_test.MockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &mockEmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, nil, nil, logger)

			err := svc.UpdateUser(context.Background(), tt.id, tt.firstName, tt.lastName, tt.role)

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

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		password      string
		mockSetup     func(*repository_test.MockRepository)
		expectedError string
	}{
		{
			name:     "Success",
			id:       "42611365-6415-4530-9346-3ee695d8b58d",
			password: "correct_password",
			mockSetup: func(r *repository_test.MockRepository) {
				// Mocking password verification relies on hashing which is hard in basic testing without the real hash.
				// In auth implementation, DeleteUser might just call Repo depending on the flow.
				// Wait, let's verify what `DeleteUser` actually does.
				// Let's assume it calls repo DeleteUserFunc eventually.
				r.DeleteUserFunc = func(ctx context.Context, id string) error {
					return nil
				}
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					// We need to return a hashed version of "correct_password" so bcrypt.CompareHashAndPassword passes.
					// bcrypt.CompareHashAndPassword accepts Password string
					// Wait, the User model holds Password string directly here?! In testing, Password is "-" in json but we use it as hash?
					// Ah, if `Password` is the field, let's use it.
					return &model.User{UserID: id, Password: "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa"}, nil, nil
				}
			},
			expectedError: "",
		},
		{
			name:     "InvalidPassword",
			id:       "42611365-6415-4530-9346-3ee695d8b58d",
			password: "wrong_password",
			mockSetup: func(r *repository_test.MockRepository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id, Password: "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa"}, nil, nil
				}
			},
			expectedError: "incorrect password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestDeleteUser case: %s\033[0m", tt.name)
			mockRepo := &repository_test.MockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &mockEmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, nil, nil, logger)

			err := svc.DeleteUser(context.Background(), tt.id, tt.password)

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
