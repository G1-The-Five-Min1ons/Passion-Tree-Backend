package learning_path_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/platform/aiclient"
)

// mockRepo for LearningPath
type mockRepo struct {
	GetAllLearningPathFunc              func(ctx context.Context) ([]model.LearningPath, error)
	GetLearningPathByIDFunc             func(ctx context.Context, path_id string) (*model.LearningPath, error)
	CreateLearningPathFunc              func(ctx context.Context, req model.CreatePathRequest) (string, error)
	UpdateLearningPathFunc              func(ctx context.Context, path_id string, req model.UpdatePathRequest) error
	DeleteLearningPathFunc              func(ctx context.Context, path_id string) error
	EnrollLearningPathUserFunc          func(ctx context.Context, pathID string, userID string) error
	GetLearningPathEnrollmentStatusFunc func(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error)
	GetUserPathProgressFunc             func(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error)
	UpdateLearningPathImageFunc         func(ctx context.Context, pathID string, coverImgURL string) error
}

func (m *mockRepo) GetAllLearningPath(ctx context.Context) ([]model.LearningPath, error) {
	if m.GetAllLearningPathFunc != nil {
		return m.GetAllLearningPathFunc(ctx)
	}
	return nil, nil
}
func (m *mockRepo) GetLearningPathByID(ctx context.Context, path_id string) (*model.LearningPath, error) {
	if m.GetLearningPathByIDFunc != nil {
		return m.GetLearningPathByIDFunc(ctx, path_id)
	}
	return nil, nil
}
func (m *mockRepo) CreateLearningPath(ctx context.Context, req model.CreatePathRequest) (string, error) {
	if m.CreateLearningPathFunc != nil {
		return m.CreateLearningPathFunc(ctx, req)
	}
	return "", nil
}
func (m *mockRepo) UpdateLearningPath(ctx context.Context, path_id string, req model.UpdatePathRequest) error {
	if m.UpdateLearningPathFunc != nil {
		return m.UpdateLearningPathFunc(ctx, path_id, req)
	}
	return nil
}
func (m *mockRepo) DeleteLearningPath(ctx context.Context, path_id string) error {
	if m.DeleteLearningPathFunc != nil {
		return m.DeleteLearningPathFunc(ctx, path_id)
	}
	return nil
}
func (m *mockRepo) EnrollLearningPathUser(ctx context.Context, pathID string, userID string) error {
	if m.EnrollLearningPathUserFunc != nil {
		return m.EnrollLearningPathUserFunc(ctx, pathID, userID)
	}
	return nil
}
func (m *mockRepo) GetLearningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error) {
	if m.GetLearningPathEnrollmentStatusFunc != nil {
		return m.GetLearningPathEnrollmentStatusFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *mockRepo) GetUserPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error) {
	if m.GetUserPathProgressFunc != nil {
		return m.GetUserPathProgressFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *mockRepo) UpdateLearningPathImage(ctx context.Context, pathID string, coverImgURL string) error {
	if m.UpdateLearningPathImageFunc != nil {
		return m.UpdateLearningPathImageFunc(ctx, pathID, coverImgURL)
	}
	return nil
}

// Implement Database interface
func (m *mockRepo) GetDB() repository.Database { return nil }

// Implement RepositoryNode
func (m *mockRepo) GetNodeByID(ctx context.Context, nodeID string) (*model.Node, error) {
	return nil, nil
}
func (m *mockRepo) CreateNode(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	return "", nil
}
func (m *mockRepo) GetNodesByPathID(ctx context.Context, pathID string) ([]model.Node, error) {
	return nil, nil
}
func (m *mockRepo) UpdateNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
	return nil
}
func (m *mockRepo) DeleteNode(ctx context.Context, nodeID string) error { return nil }
func (m *mockRepo) CreateMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error) {
	return "", nil
}
func (m *mockRepo) GetMaterialsByNodeID(ctx context.Context, nodeID string) ([]model.NodeMaterial, error) {
	return nil, nil
}
func (m *mockRepo) DeleteMaterial(ctx context.Context, materialID string) error    { return nil }
func (m *mockRepo) UpdateNodeSequence(ctx context.Context, nodeIDs []string) error { return nil }
func (m *mockRepo) CreateNodeWithContent(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	return "", nil
}

// Implement RepositoryComment
func (m *mockRepo) CreateComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	return "", nil
}
func (m *mockRepo) GetCommentsByNodeID(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	return nil, nil
}
func (m *mockRepo) DeleteComment(ctx context.Context, commentID string) error { return nil }
func (m *mockRepo) CreateReaction(ctx context.Context, req model.CreateReactionRequest) error {
	return nil
}
func (m *mockRepo) GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error) {
	return nil, nil
}
func (m *mockRepo) CreateMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	return "", nil
}

// Implement RepositoryQuiz
func (m *mockRepo) CreateQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
	return "", nil
}
func (m *mockRepo) GetQuestionsByNodeID(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
	return nil, nil
}
func (m *mockRepo) DeleteQuestion(ctx context.Context, questionID string) error { return nil }
func (m *mockRepo) CreateChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
	return "", nil
}
func (m *mockRepo) GetChoicesByQuestionID(ctx context.Context, questionID string) ([]model.QuestionChoice, error) {
	return nil, nil
}
func (m *mockRepo) DeleteChoice(ctx context.Context, choiceID string) error { return nil }

// Implement RepositoryHistory
func (m *mockRepo) GetHistoryByUserID(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
	return nil, nil
}

// Implement RepositoryResume
func (m *mockRepo) GetNextNodeID(ctx context.Context, userID string, pathID string) (string, error) {
	return "", nil
}

func TestCreatePath(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreatePathRequest
		mockSetup     func(*mockRepo)
		expectedError string
		expectedID    string
	}{
		{
			name: "Success",
			req: model.CreatePathRequest{
				Title:       "Go Lang",
				CoverImgURL: "https://example.com/learning-path/go.png",
				CreatorID:   "user-1",
			},
			mockSetup: func(m *mockRepo) {
				m.CreateLearningPathFunc = func(ctx context.Context, req model.CreatePathRequest) (string, error) {
					return "path-123", nil
				}
			},
			expectedID:    "path-123",
			expectedError: "",
		},
		{
			name: "MissingTitle",
			req: model.CreatePathRequest{
				Title:       "",
				CoverImgURL: "url",
			},
			mockSetup:     nil,
			expectedError: "title cannot be empty",
		},
		{
			name: "RepositoryError",
			req: model.CreatePathRequest{
				Title:       "Error Path",
				CoverImgURL: "url",
			},
			mockSetup: func(m *mockRepo) {
				m.CreateLearningPathFunc = func(ctx context.Context, req model.CreatePathRequest) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			aiClient := aiclient.NewAIClient("http://mock-ai")

			// Use service.NewService
			svc := service.NewService(mock, aiClient, logger)

			id, err := svc.CreatePath(context.Background(), tt.req)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if id != tt.expectedID {
					t.Errorf("Expected ID %s, got %s", tt.expectedID, id)
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

func TestGetPaths(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mockRepo)
		expectedCount int
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func(m *mockRepo) {
				m.GetAllLearningPathFunc = func(ctx context.Context) ([]model.LearningPath, error) {
					return []model.LearningPath{
						{PathID: "1", Title: "Path 1"},
						{PathID: "2", Title: "Path 2"},
					}, nil
				}
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name: "Error",
			mockSetup: func(m *mockRepo) {
				m.GetAllLearningPathFunc = func(ctx context.Context) ([]model.LearningPath, error) {
					return nil, errors.New("db fail")
				}
			},
			expectedCount: 0,
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			// Use service.NewService (nil AI client)
			svc := service.NewService(mock, nil, logger)

			paths, err := svc.GetPaths(context.Background())

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(paths) != tt.expectedCount {
					t.Errorf("Expected %d paths, got %d", tt.expectedCount, len(paths))
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

func TestGetPathDetails(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		mockSetup     func(*mockRepo)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			mockSetup: func(m *mockRepo) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return &model.LearningPath{PathID: "p1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:   "NotFound",
			pathID: "p2",
			mockSetup: func(m *mockRepo) {
				m.GetLearningPathByIDFunc = func(ctx context.Context, path_id string) (*model.LearningPath, error) {
					return nil, sql.ErrNoRows
				}
			},
			// "learning path with id '%s' not found"
			expectedError: "learning path with id 'p2' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			// Use service.NewService (nil AI client)
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetPathDetails(context.Background(), tt.pathID)
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
