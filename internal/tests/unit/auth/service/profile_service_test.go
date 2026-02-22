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

func TestUpdateProfile(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		profile       *model.Profile
		setup         func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:   "Success",
			userID: "user-1",
			profile: &model.Profile{
				Bio:      "New Bio",
				Location: "New Location",
			},
			setup: func(r *repository_test.Repository) {
				r.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, *model.Profile, error) {
					return &model.User{UserID: id}, &model.Profile{}, nil
				}
				r.UpdateProfileFunc = func(ctx context.Context, userID string, profile *model.Profile) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:   "DatabaseError",
			userID: "user-2",
			profile: &model.Profile{
				Bio: "Error Bio",
			},
			setup: func(r *repository_test.Repository) {
				r.UpdateProfileFunc = func(ctx context.Context, userID string, profile *model.Profile) error {
					return errors.New("db save failed")
				}
			},
			expectedError: "user with id 'user-2' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestUpdateProfile case: %s\033[0m", tt.name)
			mockRepo := &repository_test.Repository{}
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewUserService(mockRepo, nil, nil, nil, logger)

			err := svc.UpdateProfile(context.Background(), tt.userID, tt.profile)

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
