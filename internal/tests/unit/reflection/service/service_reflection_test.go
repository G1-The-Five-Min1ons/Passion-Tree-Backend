package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"passiontree/internal/platform/aiclient"

	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/service"
	repository_test "passiontree/internal/tests/unit/reflection/repository"
)

func TestGetReflectionByID(t *testing.T) {
	tests := []struct {
		name          string
		reflectID     string
		mockSetup     func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *repository_test.Repository) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return &model.Reflection{ReflectID: "r1"}, nil
				}
			},
			expectedError: "",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *repository_test.Repository) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "reflection with id 'r2' not found",
		},
		{
			name:      "InternalError",
			reflectID: "r3",
			mockSetup: func(m *repository_test.Repository) {
				m.GetReflectionByIDFunc = func(ctx context.Context, reflectID string) (*model.Reflection, error) {
					return nil, errors.New("db fail")
				}
			},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetReflectionByID case: %s\033[0m", tt.name)
			mock := &repository_test.Repository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

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

func TestGetAllReflections(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetAllReflectionsFunc: func(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
				return []model.Reflection{{ReflectID: "r1"}}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetAllReflections(context.Background(), model.GetReflectionsFilter{})
		if err != nil || len(res) == 0 {
			t.Errorf("Expected valid reflections list, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetAllReflectionsFunc: func(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
				return nil, errors.New("db error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetAllReflections(context.Background(), model.GetReflectionsFilter{})
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected an internal server error, got %v", err)
		}
	})
}

func TestUpdateReflection(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateReflectionFunc: func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateReflectionRequest{LearningReflect: "Test", MoodReflect: "Happy"}
		err := svc.UpdateReflection(context.Background(), "r1", req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(nil, nil, logger)

		err := svc.UpdateReflection(context.Background(), "ref-1", model.UpdateReflectionRequest{})
		if err == nil || !strings.Contains(err.Error(), "learning_reflect is required") {
			t.Errorf("Expected empty body error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateReflectionFunc: func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
				return sql.ErrNoRows
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateReflectionRequest{LearningReflect: "Test", MoodReflect: "Happy"}
		err := svc.UpdateReflection(context.Background(), "r2", req)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})

	t.Run("DuplicateKey", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateReflectionFunc: func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
				return errors.New("duplicate key error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateReflectionRequest{LearningReflect: "Test", MoodReflect: "Happy"}
		err := svc.UpdateReflection(context.Background(), "r3", req)
		if err == nil || !strings.Contains(err.Error(), "reflection with this information already exists") {
			t.Errorf("Expected duplicate key error, got %v", err)
		}
	})

	t.Run("ForeignKey", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateReflectionFunc: func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
				return errors.New("foreign key constraint error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateReflectionRequest{LearningReflect: "Test", MoodReflect: "Happy"}
		err := svc.UpdateReflection(context.Background(), "r4", req)
		if err == nil || !strings.Contains(err.Error(), "invalid tree_node_id: node does not exist") {
			t.Errorf("Expected foreign key error, got %v", err)
		}
	})

	t.Run("MissingReflectID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(nil, nil, logger)

		err := svc.UpdateReflection(context.Background(), "", model.UpdateReflectionRequest{
			LearningReflect: "Good",
		})
		if err == nil || !strings.Contains(err.Error(), "reflect_id is required") {
			t.Errorf("Expected empty reflectID error, got %v", err)
		}
	})

	t.Run("MissingMoodReflect", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(nil, nil, logger)

		err := svc.UpdateReflection(context.Background(), "r1", model.UpdateReflectionRequest{
			LearningReflect: "Good",
		})
		if err == nil || !strings.Contains(err.Error(), "mood_reflect is required") {
			t.Errorf("Expected empty mood_reflect error, got %v", err)
		}
	})
}

func TestDeleteReflection(t *testing.T) {
	tests := []struct {
		name          string
		reflectID     string
		mockSetup     func(*repository_test.Repository)
		expectedError string
	}{
		{
			name:      "Success",
			reflectID: "r1",
			mockSetup: func(m *repository_test.Repository) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "MissingReflectID",
			reflectID:     "",
			expectedError: "reflect_id is required",
		},
		{
			name:      "NotFound",
			reflectID: "r2",
			mockSetup: func(m *repository_test.Repository) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "reflection with id 'r2' not found",
		},
		{
			name:      "ForeignKeyError",
			reflectID: "r3",
			mockSetup: func(m *repository_test.Repository) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return errors.New("foreign key constraint error")
				}
			},
			expectedError: "cannot delete reflection: there are existing dependencies",
		},
		{
			name:      "InternalError",
			reflectID: "r4",
			mockSetup: func(m *repository_test.Repository) {
				m.DeleteReflectionFunc = func(ctx context.Context, reflectID string) error {
					return errors.New("db error")
				}
			},
			expectedError: "internal server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestDeleteReflection case: %s\033[0m", tt.name)
			mock := &repository_test.Repository{}
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

func TestCreateReflection(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateReflectionRequest
		mockSetup     func(*repository_test.Repository)
		aiMockStatus  int
		aiMockBody    interface{}
		noAIClient    bool
		expectedError string
	}{
		{
			name: "Success",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good",
				MoodReflect:     "Happy",
				TreeNodeID:      "t1",
			},
			mockSetup: func(m *repository_test.Repository) {
				m.CreateReflectionFunc = func(ctx context.Context, req model.CreateReflectionRequest, summary, sentiment string, emotion *string, struggle string, aiScore, reflectScore, weightedScore float64) (string, error) {
					return "r1", nil
				}
			},
			aiMockStatus: http.StatusOK,
			aiMockBody: aiclient.SentimentResponse{
				Summary: "Sum", SentimentAnalysis: "Sent", PrimaryEmotion: func() *string { s := "Happy"; return &s }(), StrugglePoint: "None", DevelopmentPlan: []string{"Plan"},
				AIConfidentScore: 90, ReflectionScore: 80, WeightedReflectionScore: 85,
			},
			expectedError: "",
		},
		{
			name: "MissingLearningReflect",
			req: model.CreateReflectionRequest{
				MoodReflect: "Happy", TreeNodeID: "t1",
			},
			expectedError: "learning_reflect is required",
		},
		{
			name: "MissingMoodReflect",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good", TreeNodeID: "t1",
			},
			expectedError: "mood_reflect is required",
		},
		{
			name: "MissingTreeNodeID",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good", MoodReflect: "Happy",
			},
			expectedError: "tree_node_id is required",
		},
		{
			name: "NoAIConfigured",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good", MoodReflect: "Happy", TreeNodeID: "t1",
			},
			noAIClient:    true,
			expectedError: "internal server error",
		},
		{
			name: "AIFailedAnalysis",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good", MoodReflect: "Happy", TreeNodeID: "t1",
			},
			aiMockStatus:  http.StatusInternalServerError,
			expectedError: "internal server error",
		},
		{
			name: "DuplicateKeyDB",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good", MoodReflect: "Happy", TreeNodeID: "t1",
			},
			mockSetup: func(m *repository_test.Repository) {
				m.CreateReflectionFunc = func(ctx context.Context, req model.CreateReflectionRequest, summary, sentiment string, emotion *string, struggle string, aiScore, reflectScore, weightedScore float64) (string, error) {
					return "", errors.New("duplicate key error")
				}
			},
			aiMockStatus:  http.StatusOK,
			aiMockBody:    aiclient.SentimentResponse{},
			expectedError: "reflection with this ID already exists",
		},
		{
			name: "ForeignKeyDB",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good", MoodReflect: "Happy", TreeNodeID: "t1",
			},
			mockSetup: func(m *repository_test.Repository) {
				m.CreateReflectionFunc = func(ctx context.Context, req model.CreateReflectionRequest, summary, sentiment string, emotion *string, struggle string, aiScore, reflectScore, weightedScore float64) (string, error) {
					return "", errors.New("foreign key constraint error")
				}
			},
			aiMockStatus:  http.StatusOK,
			aiMockBody:    aiclient.SentimentResponse{},
			expectedError: "invalid tree_node_id or user_id",
		},
		{
			name: "InternalErrorDB",
			req: model.CreateReflectionRequest{
				LearningReflect: "Good", MoodReflect: "Happy", TreeNodeID: "t1",
			},
			mockSetup: func(m *repository_test.Repository) {
				m.CreateReflectionFunc = func(ctx context.Context, req model.CreateReflectionRequest, summary, sentiment string, emotion *string, struggle string, aiScore, reflectScore, weightedScore float64) (string, error) {
					return "", errors.New("db error")
				}
			},
			aiMockStatus:  http.StatusOK,
			aiMockBody:    aiclient.SentimentResponse{},
			expectedError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestCreateReflection case: %s\033[0m", tt.name)

			// Start local HTTP server for AI mock
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.aiMockStatus)
				if tt.aiMockBody != nil {
					json.NewEncoder(w).Encode(tt.aiMockBody)
				}
			}))
			defer ts.Close()

			mock := &repository_test.Repository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			var ai *aiclient.AIClient
			if !tt.noAIClient {
				ai = aiclient.NewAIClient(ts.URL)
			}

			svc := service.NewService(mock, ai, logger)

			_, err := svc.CreateReflection(context.Background(), tt.req)
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
