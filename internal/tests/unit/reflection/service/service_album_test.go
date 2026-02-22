package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/service"
	repository_test "passiontree/internal/tests/unit/reflection/repository"
)

func TestCreateAlbum(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateAlbumRequest
		mockSetup     func(*repository_test.MockRefRepo)
		expectedError string
	}{
		{
			name: "Success",
			req: model.CreateAlbumRequest{
				AlbumName: "My Album",
				UserID:    "user-1",
			},
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.CreateAlbumFunc = func(ctx context.Context, req model.CreateAlbumRequest) (string, error) {
					return "album1", nil
				}
				m.GetAlbumByIDFunc = func(ctx context.Context, albumID string) (*model.Album, error) {
					return &model.Album{AlbumID: albumID}, nil
				}
			},
			expectedError: "",
		},
		{
			name: "MissingAlbumName",
			req: model.CreateAlbumRequest{
				UserID: "user-1",
			},
			mockSetup:     func(*repository_test.MockRefRepo) {},
			expectedError: "album_name is required",
		},
		{
			name: "MissingUserID",
			req: model.CreateAlbumRequest{
				AlbumName: "My Album",
			},
			mockSetup:     func(*repository_test.MockRefRepo) {},
			expectedError: "user_id is required",
		},
		{
			name: "InternalError",
			req: model.CreateAlbumRequest{
				AlbumName: "My Album",
				UserID:    "user-1",
			},
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.CreateAlbumFunc = func(ctx context.Context, req model.CreateAlbumRequest) (string, error) {
					return "", errors.New("db error")
				}
				m.GetAlbumByIDFunc = func(ctx context.Context, albumID string) (*model.Album, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestCreateAlbum case: %s\033[0m", tt.name)
			mock := &repository_test.MockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.CreateAlbum(context.Background(), tt.req)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGetAlbumByID(t *testing.T) {
	tests := []struct {
		name          string
		albumID       string
		mockSetup     func(*repository_test.MockRefRepo)
		expectedError string
	}{
		{
			name:    "Success",
			albumID: "album1",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.GetAlbumByIDFunc = func(ctx context.Context, albumID string) (*model.Album, error) {
					return &model.Album{AlbumID: "album1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyID",
			albumID:       "",
			mockSetup:     func(*repository_test.MockRefRepo) {},
			expectedError: "album_id is required",
		},
		{
			name:    "NotFound",
			albumID: "album2",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.GetAlbumByIDFunc = func(ctx context.Context, albumID string) (*model.Album, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "album with id 'album2' not found",
		},
		{
			name:    "DatabaseError",
			albumID: "album3",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.GetAlbumByIDFunc = func(ctx context.Context, albumID string) (*model.Album, error) {
					return nil, errors.New("db failure")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetAlbumByID case: %s\033[0m", tt.name)
			mock := &repository_test.MockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetAlbumByID(context.Background(), tt.albumID)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestGetAlbumsByUserID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*repository_test.MockRefRepo)
		expectedError string
	}{
		{
			name:   "Success",
			userID: "user1",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.GetAlbumsByUserIDFunc = func(ctx context.Context, userID string) ([]model.Album, error) {
					return []model.Album{{AlbumID: "a1"}}, nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyUserID",
			userID:        "",
			mockSetup:     func(*repository_test.MockRefRepo) {},
			expectedError: "user_id is required",
		},
		{
			name:   "DatabaseError",
			userID: "user2",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.GetAlbumsByUserIDFunc = func(ctx context.Context, userID string) ([]model.Album, error) {
					return nil, errors.New("db fail")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetAlbumsByUserID case: %s\033[0m", tt.name)
			mock := &repository_test.MockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetAlbumsByUserID(context.Background(), tt.userID)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestUpdateAlbum(t *testing.T) {
	tests := []struct {
		name          string
		albumID       string
		req           model.UpdateAlbumRequest
		mockSetup     func(*repository_test.MockRefRepo)
		expectedError string
	}{
		{
			name:    "Success",
			albumID: "album1",
			req:     model.UpdateAlbumRequest{AlbumName: "new"},
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.UpdateAlbumFunc = func(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyBody",
			albumID:       "album1",
			req:           model.UpdateAlbumRequest{},
			mockSetup:     func(*repository_test.MockRefRepo) {},
			expectedError: "album_name is required",
		},
		{
			name:          "EmptyAlbumID",
			albumID:       "",
			req:           model.UpdateAlbumRequest{AlbumName: "new"},
			mockSetup:     func(*repository_test.MockRefRepo) {},
			expectedError: "album_id is required",
		},
		{
			name:    "NotFound",
			albumID: "album2",
			req:     model.UpdateAlbumRequest{AlbumName: "new"},
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.UpdateAlbumFunc = func(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "album with id 'album2' not found",
		},
		{
			name:    "InternalError",
			albumID: "album3",
			req:     model.UpdateAlbumRequest{AlbumName: "new"},
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.UpdateAlbumFunc = func(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
					return errors.New("db err")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestUpdateAlbum case: %s\033[0m", tt.name)
			mock := &repository_test.MockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.UpdateAlbum(context.Background(), tt.albumID, tt.req)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestDeleteAlbum(t *testing.T) {
	tests := []struct {
		name          string
		albumID       string
		mockSetup     func(*repository_test.MockRefRepo)
		expectedError string
	}{
		{
			name:    "Success",
			albumID: "album1",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.DeleteAlbumFunc = func(ctx context.Context, albumID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyID",
			albumID:       "",
			mockSetup:     func(*repository_test.MockRefRepo) {},
			expectedError: "album_id is required",
		},
		{
			name:    "NotFound",
			albumID: "album2",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.DeleteAlbumFunc = func(ctx context.Context, albumID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "album with id 'album2' not found",
		},
		{
			name:    "DatabaseError",
			albumID: "album3",
			mockSetup: func(m *repository_test.MockRefRepo) {
				m.DeleteAlbumFunc = func(ctx context.Context, albumID string) error {
					return errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestDeleteAlbum case: %s\033[0m", tt.name)
			mock := &repository_test.MockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.DeleteAlbum(context.Background(), tt.albumID)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}
