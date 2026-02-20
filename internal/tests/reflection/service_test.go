package reflection_test

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
)

// mockRefRepo implements repository.RepositoryReflection
type mockRefRepo struct {
	GetReflectionByIDFunc func(ctx context.Context, reflectID string) (*model.Reflection, error)
	UpdateReflectionFunc  func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflectionFunc  func(ctx context.Context, reflectID string) error
	GetAllReflectionsFunc func(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error)
	CreateReflectionFunc  func(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error)
}

func (m *mockRefRepo) GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error) {
	if m.GetReflectionByIDFunc != nil {
		return m.GetReflectionByIDFunc(ctx, reflectID)
	}
	return nil, nil
}
func (m *mockRefRepo) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
	if m.UpdateReflectionFunc != nil {
		return m.UpdateReflectionFunc(ctx, reflectID, req)
	}
	return nil
}
func (m *mockRefRepo) DeleteReflection(ctx context.Context, reflectID string) error {
	if m.DeleteReflectionFunc != nil {
		return m.DeleteReflectionFunc(ctx, reflectID)
	}
	return nil
}
func (m *mockRefRepo) GetAllReflections(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
	if m.GetAllReflectionsFunc != nil {
		return m.GetAllReflectionsFunc(ctx, filter)
	}
	return nil, nil
}
func (m *mockRefRepo) CreateReflection(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error) {
	if m.CreateReflectionFunc != nil {
		return m.CreateReflectionFunc(ctx, req, summary, sentimentAnalysis, primaryEmotion, strugglePoint, aiConfidentScore, reflectionScore, weightedReflectionScore)
	}
	return "", nil
}

// Implement other interface methods as no-ops
func (m *mockRefRepo) CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (string, error) {
	return "", nil
}
func (m *mockRefRepo) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	return nil, nil
}
func (m *mockRefRepo) GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error) {
	return nil, nil
}
func (m *mockRefRepo) UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
	return nil
}
func (m *mockRefRepo) DeleteAlbum(ctx context.Context, albumID string) error { return nil }
func (m *mockRefRepo) CreateTree(ctx context.Context, req model.CreateTreeRequest) (string, error) {
	return "", nil
}
func (m *mockRefRepo) GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error) {
	return nil, nil
}
func (m *mockRefRepo) GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error) {
	return nil, nil
}
func (m *mockRefRepo) GetTreesWithNodesByAlbumID(ctx context.Context, albumID string) ([]model.TreeResponse, error) {
	return nil, nil
}
func (m *mockRefRepo) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
	return nil
}
func (m *mockRefRepo) DeleteTree(ctx context.Context, treeID string) error              { return nil }
func (m *mockRefRepo) PauseTree(ctx context.Context, treeID string, isPause bool) error { return nil }
func (m *mockRefRepo) AddSingleTreeNode(ctx context.Context, req model.CreateTreeNodeRequest) (string, error) {
	return "", nil
}
func (m *mockRefRepo) GetTreeNodesByTreeID(ctx context.Context, treeID string) ([]model.TreeNode, error) {
	return nil, nil
}
func (m *mockRefRepo) GetTreeNodeByID(ctx context.Context, treeNodeID string) (*model.TreeNode, error) {
	return nil, nil
}
func (m *mockRefRepo) UpdateTreeNode(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error {
	return nil
}
func (m *mockRefRepo) DeleteTreeNode(ctx context.Context, treeNodeID string) error { return nil }
func (m *mockRefRepo) CreateTreeNodes(ctx context.Context, treeID string, pathID string) error {
	return nil
}
func (m *mockRefRepo) GetNodesByPathID(ctx context.Context, pathID string) ([]model.TreeNode, error) {
	return nil, nil
}

func TestGetReflectionByID(t *testing.T) {
	tests := []struct {
		name          string
		reflectID     string
		mockSetup     func(*mockRefRepo)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *mockRefRepo) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return &model.Reflection{ReflectID: "r1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *mockRefRepo) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "reflection with id 'r2' not found",
		},
		{
			name:      "InternalError",
			reflectID: "r3",
			mockSetup: func(m *mockRefRepo) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return nil, errors.New("db fail")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger) // aiClient nil

			_, err := svc.GetReflectionByID(context.Background(), tt.reflectID)

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

func TestDeleteReflection(t *testing.T) {
	tests := []struct {
		name          string
		reflectID     string
		mockSetup     func(*mockRefRepo)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *mockRefRepo) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *mockRefRepo) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "reflection with id 'r2' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRefRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.DeleteReflection(context.Background(), tt.reflectID)

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

func TestCreateReflection_NoAI(t *testing.T) {
	// Test that it fails when AI service is not configured (nil aiClient)
	mock := &mockRefRepo{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewService(mock, nil, logger)

	req := model.CreateReflectionRequest{
		LearningReflect: "I learned a lot",
		MoodReflect:     "Happy",
		TreeNodeID:      "node-1",
	}

	_, err := svc.CreateReflection(context.Background(), req)
	expectedError := "internal server error"
	if err == nil {
		t.Errorf("Expected error '%s', got nil", expectedError)
	} else if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCreateReflection_MissingFields(t *testing.T) {
	mock := &mockRefRepo{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewService(mock, nil, logger)

	req := model.CreateReflectionRequest{
		// Missing fields
	}

	_, err := svc.CreateReflection(context.Background(), req)
	expectedError := "learning_reflect is required"
	if err == nil {
		t.Errorf("Expected error '%s', got nil", expectedError)
	} else if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}
