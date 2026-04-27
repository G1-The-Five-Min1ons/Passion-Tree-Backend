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
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:          "MissingEmail",
			user:          &model.User{Username: "farloss", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile:       &model.Profile{},
			setup:         nil,
			expectedError: "email is required",
		},
		{
			name:          "MissingPassword",
			user:          &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Role: "student"},
			profile:       &model.Profile{},
			setup:         nil,
			expectedError: "password is required",
		},
		{
			name:          "ShortPassword",
			user:          &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "short", Role: "student"},
			profile:       &model.Profile{},
			setup:         nil,
			expectedError: "password must be at least 8 characters long",
		},
		{
			name:          "MissingUsername",
			user:          &model.User{Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile:       &model.Profile{},
			setup:         nil,
			expectedError: "username is required",
		},
		{
			name:          "InvalidEmailFormat",
			user:          &model.User{Username: "farloss", Email: "invalid-email", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile:       &model.Profile{},
			setup:         nil,
			expectedError: "invalid email format",
		},
		{
			name:    "MissingRole",
			user:    &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword"},
			profile: &model.Profile{},
			setup:   nil,
			expectedError: "role is required",
		},
		{
			name:    "RoleInvalid",
			user:    &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "invalidrole"},
			profile: &model.Profile{},
			setup:   nil,
			expectedError: "role must be one of 'student', 'teacher', 'admin', 'pending', 'user', or 'moderator'",
		},
		{
			name:    "RegisterSuccess",
			user:    &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile: &model.Profile{},
			setup: func(r *repository_test.Repository) {
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
			setup: func(r *repository_test.Repository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{UserID: "existing-id"}, nil // Email is taken
				}
			},
			expectedError: "email already registered",
		},
		{
			name:    "UsernameAlreadyExists",
			user:    &model.User{Username: "farloss", Email: "thirapat.pant@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile: &model.Profile{},
			setup: func(r *repository_test.Repository) {
				r.GetUserByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return &model.User{UserID: "existing-id"}, nil // Username is taken
				}
			},
			expectedError: "username already registered",
		},
		{
			name: "SaveTokenError",
			user: &model.User{Username: "farloss", Email: "thirapatth@gmail.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile: &model.Profile{},
			setup: func(r *repository_test.Repository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, nil // Ensure email is not taken
				}
				r.CreateUserFunc = func(ctx context.Context, user *model.User, profile *model.Profile) (string, error) {
					return "new-user-id", nil
				}
				r.CreateTokenFunc = func(ctx context.Context, token *model.Token) error {
					return errors.New("failed to save token")
				}
			},
			expectedError: "internal server error",
		},
		{
			name:    "DatabaseError",
			user:    &model.User{Username: "farloss", Email: "error@example.com", FirstName: "Thiraphat", LastName: "Panthong", Password: "securepassword", Role: "student"},
			profile: &model.Profile{},
			setup: func(r *repository_test.Repository) {
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
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &EmailService{}
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
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:          "MissingUserID",
			id:            "",
			setup:         nil,
			expectedError: "user_id is required",
		},
		{
			name: "GetUserByIDSuccess",
			id:   "user-1",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id, FirstName: "Test"}, &model.Profile{}, nil
				}
			},
			expectedError: "",
		},
		{
			name: "GetUserByIDNotFound",
			id:   "user-2",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return nil, nil, nil
				}
			},
			expectedError: "user with id 'user-2' not found",
		},
		{
			name: "GetUserByIDRepositoryError",
			id:   "user-999",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return nil, nil, errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetUserByID case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &EmailService{}
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
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:          "MissingEmail",
			email:         "",
			setup:         nil,
			expectedError: "email is required",
		},
		{
			name:  "GetUserByEmailSuccess",
			email: "found@example.com",
			setup: func(r *repository_test.Repository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return &model.User{Email: email}, nil
				}
			},
			expectedError: "",
		},
		{
			name:  "GetUserByEmailNotFound",
			email: "missing@example.com",
			setup: func(r *repository_test.Repository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, nil
				}
			},
			expectedError: "user with email 'missing@example.com' not found",
		},
		{
			name:  "GetUserByEmailRepositoryError",
			email: "error@example.com",
			setup: func(r *repository_test.Repository) {
				r.GetUserByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetUserByEmail case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &EmailService{}
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
		username      string
		firstName     string
		lastName      string
		role          string
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:          "MissingUserID",
			id:            "",
			setup:         nil,
			firstName:     "Updated",
			expectedError: "user_id is required",
		},
		{
			name:      "Success",
			id:        "user-1",
			username:  "farloss_new",
			firstName: "Thiraphat",
            lastName:  "Panthong",
			role:      "student",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, &model.Profile{}, nil
				}
				r.UpdateUserFunc = func(ctx context.Context, id string, username string, fName string, lName string, role string) error {
                    if fName != "Thiraphat" || lName != "Panthong" {
                        return errors.New("wrong data passed to repository")
                    }
                    return nil
                }
            },
            expectedError: "",
        },
		{
			name: "UserNotFound",
			id:   "user-2",
			setup: func(r *repository_test.Repository) {	
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return nil, nil, nil
				}
			},
			expectedError: "user with id 'user-2' not found",
		},
			{
				name: "UsernameAlreadyExists",
				id:   "user-3",
				firstName: "Existing",
				setup: func(r *repository_test.Repository) {
					r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
						return &model.User{UserID: id}, &model.Profile{}, nil
					}
					r.UpdateUserFunc = func(ctx context.Context, id string, username string, firstName string, lastName string, role string) error {
						return errors.New("username already exists")
					}
				},
				expectedError: "username already exists",
			},
		{
			name: "RepositoryError",
			id:   "user-2",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, &model.Profile{}, nil
				}
				r.UpdateUserFunc = func(ctx context.Context, id string, username string, firstName string, lastName string, role string) error {
					return errors.New("db update failed")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestUpdateUser case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &EmailService{}
			svc := service.NewUserService(mockRepo, mockEmailSvc, nil, nil, logger)

			err := svc.UpdateUser(context.Background(), tt.id, tt.username, tt.firstName, tt.lastName, tt.role)

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
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:          "MissingUserID",
			id:            "",
			password:	  "any_password",
			setup:         nil,
			expectedError: "user_id is required",
		},
		{
			name:          "MissingPassword",
			id:            "user-123",
			password:      "",
			setup:         nil,
			expectedError: "password is required",
		},
		{
			name:          "UserNotFound",
			id:            "nonexistent-user",
			password:      "any_password",
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return nil, nil, errors.New("user not found")
				}
			},
			expectedError: "user not found",	
		},
		{
			name:     "Success",
			id:       "42611365-6415-4530-9346-3ee695d8b58d",
			password: "correct_password",
			setup: func(r *repository_test.Repository) {
				r.DeleteUserFunc = func(ctx context.Context, id string) error {
					return nil
				}
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id, Password: "$2a$10$vU1OjvFQhoRzw3MTZ9uNPejompP6k6I3I4YaAYQ3AKm43B5C5AbFa"}, nil, nil
				}
			},
			expectedError: "",
		},
		{
			name:     "InvalidPassword",
			id:       "42611365-6415-4530-9346-3ee695d8b58d",
			password: "wrong_password",
			setup: func(r *repository_test.Repository) {
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
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockEmailSvc := &EmailService{}
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
