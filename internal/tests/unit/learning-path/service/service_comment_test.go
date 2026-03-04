package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/service"
	repository_test "passiontree/internal/tests/unit/learning-path/repository"
)

func TestAddComment(t *testing.T) {
	parentID := "p1"
	tests := []struct {
		name          string
		req           model.CreateCommentRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name: "Success - no parent",
			req:  model.CreateCommentRequest{Message: "Great!"},
			setup: func(m *repository_test.Repopository) {
				m.CreateCommentFunc = func(ctx context.Context, req model.CreateCommentRequest) (string, error) {
					return "c1", nil
				}
			},
			expectedError: "",
		},
		{
			name: "Success - reply with auto-mention",
			req: model.CreateCommentRequest{
				Message:  "Replying!",
				UserID:   "user-2",
				ParentID: &parentID,
			},
			setup: func(m *repository_test.Repopository) {
				m.GetCommentOwnerFunc = func(ctx context.Context, commentID string) (string, error) {
					return "user-1", nil // different owner → mention should be triggered
				}
				m.CreateCommentFunc = func(ctx context.Context, req model.CreateCommentRequest) (string, error) {
					return "c2", nil
				}
				m.CreateMentionFunc = func(ctx context.Context, req model.CreateMentionRequest) (string, error) {
					return "m1", nil
				}
			},
			expectedError: "",
		},
		{
			name: "Error - parent not found",
			req: model.CreateCommentRequest{
				Message:  "Replying!",
				UserID:   "user-2",
				ParentID: &parentID,
			},
			setup: func(m *repository_test.Repopository) {
				m.GetCommentOwnerFunc = func(ctx context.Context, commentID string) (string, error) {
					return "", errors.New("not found")
				}
			},
			expectedError: "parent comment not found",
		},
		{
			name: "Error - db error on create",
			req:  model.CreateCommentRequest{Message: "Great!"},
			setup: func(m *repository_test.Repopository) {
				m.CreateCommentFunc = func(ctx context.Context, req model.CreateCommentRequest) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestAddComment case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.AddComment(context.Background(), tt.req)
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

func TestGetNodeComments(t *testing.T) {
	tests := []struct {
		name          string
		nodeID        string
		setup         func(*repository_test.Repopository)
		expectedLen   int
		expectedError string
	}{
		{
			name:   "Success",
			nodeID: "n1",
			setup: func(m *repository_test.Repopository) {
				m.GetCommentsByNodeIDFunc = func(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
					return []model.NodeComment{{CommentID: "c1"}}, nil
				}
			},
			expectedLen:   1,
			expectedError: "",
		},
		{
			name:   "Error",
			nodeID: "n1",
			setup: func(m *repository_test.Repopository) {
				m.GetCommentsByNodeIDFunc = func(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
					return nil, errors.New("db error")
				}
			},
			expectedLen:   0,
			expectedError: "db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetNodeComments case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			comments, err := svc.GetNodeComments(context.Background(), tt.nodeID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(comments) != tt.expectedLen {
					t.Errorf("Expected %d comments, got %d", tt.expectedLen, len(comments))
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestRemoveComment(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		commentID     string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:      "Success",
			userID:    "user-1",
			commentID: "c1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteCommentFunc = func(ctx context.Context, commentID, userID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:      "Error",
			userID:    "user-1",
			commentID: "c1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteCommentFunc = func(ctx context.Context, commentID, userID string) error {
					return errors.New("db error")
				}
			},
			expectedError: "db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRemoveComment case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.RemoveComment(context.Background(), tt.userID, tt.commentID)
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

func TestUpdateComment(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		commentID      string
		message        string
		setup          func(*repository_test.Repopository)
		expectedError  string
	}{
		{
			name:      "Success - comment updated",
			userID:    "user-1",
			commentID: "c1",
			message:   "edited content",
			setup: func(m *repository_test.Repopository) {
				m.UpdateCommentFunc = func(ctx context.Context, userID, messageID, message string) (bool, error) {
					return true, nil
				}
			},
			expectedError:  "",
		},
		{
			name:      "Success - not owner, no update",
			userID:    "user-2",
			commentID: "c1",
			message:   "edited content",
			setup: func(m *repository_test.Repopository) {
				m.UpdateCommentFunc = func(ctx context.Context, userID, messageID, message string) (bool, error) {
					return false, nil
				}
			},
			expectedError:  "comment not found or not owned by you",
		},
		{
			name:      "Error",
			userID:    "user-1",
			commentID: "c1",
			message:   "edited content",
			setup: func(m *repository_test.Repopository) {
				m.UpdateCommentFunc = func(ctx context.Context, userID, messageID, message string) (bool, error) {
					return false, errors.New("db error")
				}
			},
			expectedError:  "db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestUpdateComment case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.UpdateComment(context.Background(), tt.userID, tt.commentID, tt.message)
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

func TestAddReaction(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateReactionRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name: "Success",
			req:  model.CreateReactionRequest{ReactionType: "like"},
			setup: func(m *repository_test.Repopository) {
				m.CreateReactionFunc = func(ctx context.Context, req model.CreateReactionRequest) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name: "Error",
			req:  model.CreateReactionRequest{ReactionType: "like"},
			setup: func(m *repository_test.Repopository) {
				m.CreateReactionFunc = func(ctx context.Context, req model.CreateReactionRequest) error {
					return errors.New("db error")
				}
			},
			expectedError: "db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestAddReaction case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.AddReaction(context.Background(), tt.req)
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

func TestAddMention(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateMentionRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name: "Success",
			req:  model.CreateMentionRequest{CommentID: "c1"},
			setup: func(m *repository_test.Repopository) {
				m.CreateMentionFunc = func(ctx context.Context, req model.CreateMentionRequest) (string, error) {
					return "m1", nil
				}
			},
			expectedError: "",
		},
		{
			name: "Error",
			req:  model.CreateMentionRequest{CommentID: "c1"},
			setup: func(m *repository_test.Repopository) {
				m.CreateMentionFunc = func(ctx context.Context, req model.CreateMentionRequest) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestAddMention case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.AddMention(context.Background(), tt.req)
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
