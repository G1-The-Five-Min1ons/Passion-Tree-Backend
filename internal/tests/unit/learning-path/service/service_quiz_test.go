package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/pkg/apperror"
	repository_test "passiontree/internal/tests/unit/learning-path/repository"
)

func TestAddQuestion(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateQuestionRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name: "Success",
			req: model.CreateQuestionRequest{
				QuestionText: "What is Go?",
				Type:         "multiple-choice",
				NodeID:       "n1",
			},
			setup: func(m *repository_test.Repopository) {
				m.CreateQuestionFunc = func(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
					return "q1", nil
				}
			},
			expectedError: "",
		},
		{
			name: "EmptyText",
			req: model.CreateQuestionRequest{
				QuestionText: "",
				Type:         "multiple-choice",
			},
			setup:         nil,
			expectedError: "question text is required",
		},
		{
			name: "EmptyType",
			req: model.CreateQuestionRequest{
				QuestionText: "What is Go?",
				Type:         "",
			},
			setup:         nil,
			expectedError: "question type is required",
		},
		{
			name: "DuplicateKey",
			req: model.CreateQuestionRequest{
				QuestionText: "What is Go?",
				Type:         "multiple-choice",
			},
			setup: func(m *repository_test.Repopository) {
				m.CreateQuestionFunc = func(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
					return "", apperror.NewConflict("duplicate key")
				}
			},
			expectedError: "question with this ID already exists",
		},
		{
			name: "ForeignKeyError",
			req: model.CreateQuestionRequest{
				QuestionText: "What is Go?",
				Type:         "multiple-choice",
			},
			setup: func(m *repository_test.Repopository) {
				m.CreateQuestionFunc = func(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
					return "", apperror.NewBadRequest("foreign key constraint")
				}
			},
			expectedError: "invalid node_id: node does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestAddQuestion case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.AddQuestion(context.Background(), tt.req)
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

func TestGetQuestions(t *testing.T) {
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
				m.GetQuestionsByNodeIDFunc = func(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
					return []model.NodeQuestion{{QuestionID: "q1"}}, nil
				}
			},
			expectedLen:   1,
			expectedError: "",
		},
		{
			name:          "EmptyNodeID",
			nodeID:        "",
			setup:         nil,
			expectedError: "node_id is required",
		},
		{
			name:   "Error",
			nodeID: "n1",
			setup: func(m *repository_test.Repopository) {
				m.GetQuestionsByNodeIDFunc = func(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
					return nil, errors.New("db error")
				}
			},
			expectedLen:   0,
			expectedError: "internal server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetQuestions case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			questions, err := svc.GetQuestions(context.Background(), tt.nodeID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(questions) != tt.expectedLen {
					t.Errorf("Expected %d questions, got %d", tt.expectedLen, len(questions))
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestRemoveQuestion(t *testing.T) {
	tests := []struct {
		name          string
		questionID    string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:       "Success",
			questionID: "q1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteQuestionFunc = func(ctx context.Context, questionID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyQuestionID",
			questionID:    "",
			setup:         nil,
			expectedError: "question_id is required",
		},
		{
			name:       "NotFound",
			questionID: "q2",
			setup: func(m *repository_test.Repopository) {
				m.DeleteQuestionFunc = func(ctx context.Context, questionID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "cannot delete: question id 'q2' not found",
		},
		{
			name:       "ForeignKeyError",
			questionID: "q3",
			setup: func(m *repository_test.Repopository) {
				m.DeleteQuestionFunc = func(ctx context.Context, questionID string) error {
					return apperror.NewBadRequest("foreign key constraint")
				}
			},
			expectedError: "cannot delete question: there are existing choices",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRemoveQuestion case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.RemoveQuestion(context.Background(), tt.questionID)
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

func TestAddChoice(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateChoiceRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name: "Success",
			req:  model.CreateChoiceRequest{ChoiceText: "Option A", QuestionID: "q1"},
			setup: func(m *repository_test.Repopository) {
				m.CreateChoiceFunc = func(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
					return "c1", nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyChoice",
			req:           model.CreateChoiceRequest{ChoiceText: ""},
			setup:         nil,
			expectedError: "choice text is required",
		},
		{
			name: "DuplicateKey",
			req:  model.CreateChoiceRequest{ChoiceText: "Option B"},
			setup: func(m *repository_test.Repopository) {
				m.CreateChoiceFunc = func(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
					return "", apperror.NewConflict("duplicate key")
				}
			},
			expectedError: "choice with this ID already exists",
		},
		{
			name: "ForeignKeyError",
			req:  model.CreateChoiceRequest{ChoiceText: "Option C"},
			setup: func(m *repository_test.Repopository) {
				m.CreateChoiceFunc = func(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
					return "", apperror.NewBadRequest("foreign key constraint")
				}
			},
			expectedError: "invalid question_id: question does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestAddChoice case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.AddChoice(context.Background(), tt.req)
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

func TestRemoveChoice(t *testing.T) {
	tests := []struct {
		name          string
		choiceID      string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:     "Success",
			choiceID: "c1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteChoiceFunc = func(ctx context.Context, choiceID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyChoiceID",
			choiceID:      "",
			setup:         nil,
			expectedError: "choice_id is required",
		},
		{
			name:     "NotFound",
			choiceID: "c2",
			setup: func(m *repository_test.Repopository) {
				m.DeleteChoiceFunc = func(ctx context.Context, choiceID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "cannot delete: choice id 'c2' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRemoveChoice case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.RemoveChoice(context.Background(), tt.choiceID)
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
