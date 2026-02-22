package service_test

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/pkg/apperror"
	repository_test "passiontree/internal/tests/unit/learning-path/repository"
)

func TestAddNode(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateNodeRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name: "Success",
			req: model.CreateNodeRequest{
				Title:  "Introduction",
				PathID: "p1",
			},
			setup: func(m *repository_test.Repopository) {
				m.CreateNodeWithContentFunc = func(ctx context.Context, req model.CreateNodeRequest) (string, error) {
					return "n1", nil
				}
			},
			expectedError: "",
		},
		{
			name: "EmptyTitle",
			req: model.CreateNodeRequest{
				Title:  "",
				PathID: "p1",
			},
			setup:         nil,
			expectedError: "node title is required",
		},
		{
			name: "DuplicateKey",
			req: model.CreateNodeRequest{
				Title:  "Intro",
				PathID: "p1",
			},
			setup: func(m *repository_test.Repopository) {
				m.CreateNodeWithContentFunc = func(ctx context.Context, req model.CreateNodeRequest) (string, error) {
					return "", apperror.NewConflict("duplicate key")
				}
			},
			expectedError: "node with this ID already exists",
		},
		{
			name: "ForeignKeyError",
			req: model.CreateNodeRequest{
				Title:  "Intro",
				PathID: "p1",
			},
			setup: func(m *repository_test.Repopository) {
				m.CreateNodeWithContentFunc = func(ctx context.Context, req model.CreateNodeRequest) (string, error) {
					return "", apperror.NewBadRequest("foreign key constraint")
				}
			},
			expectedError: "invalid path_id: learning path does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestAddNode case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.AddNode(context.Background(), tt.req)
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

func TestEditNode(t *testing.T) {
	tests := []struct {
		name          string
		nodeID        string
		req           model.UpdateNodeRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			nodeID: "n1",
			req:    model.UpdateNodeRequest{Title: "Updated Intro"},
			setup: func(m *repository_test.Repopository) {
				m.UpdateNodeFunc = func(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyNodeID",
			nodeID:        "",
			req:           model.UpdateNodeRequest{Title: "Updated"},
			setup:         nil,
			expectedError: "node_id is required",
		},
		{
			name:          "EmptyRequest",
			nodeID:        "n1",
			req:           model.UpdateNodeRequest{},
			setup:         nil,
			expectedError: "at least one field",
		},
		{
			name:   "NotFound",
			nodeID: "n1",
			req:    model.UpdateNodeRequest{Title: "Updated"},
			setup: func(m *repository_test.Repopository) {
				m.UpdateNodeFunc = func(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "cannot update: node id 'n1' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestEditNode case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.EditNode(context.Background(), tt.nodeID, tt.req)
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

func TestRemoveNode(t *testing.T) {
	tests := []struct {
		name          string
		nodeID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			nodeID: "n1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteNodeFunc = func(ctx context.Context, nodeID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyNodeID",
			nodeID:        "",
			setup:         nil,
			expectedError: "node_id is required",
		},
		{
			name:   "NotFound",
			nodeID: "n1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteNodeFunc = func(ctx context.Context, nodeID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "cannot delete: node id 'n1' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRemoveNode case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.RemoveNode(context.Background(), tt.nodeID)
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

func TestAddMaterial(t *testing.T) {
	tests := []struct {
		name          string
		req           model.CreateMaterialRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name: "Success",
			req: model.CreateMaterialRequest{
				Type: "video",
				URL:  "http://example.com/vid",
			},
			setup: func(m *repository_test.Repopository) {
				m.CreateMaterialFunc = func(ctx context.Context, req model.CreateMaterialRequest) (string, error) {
					return "req1", nil
				}
			},
			expectedError: "",
		},
		{
			name: "EmptyFields",
			req: model.CreateMaterialRequest{
				Type: "",
				URL:  "",
			},
			setup:         nil,
			expectedError: "material type and url are required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestAddMaterial case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.AddMaterial(context.Background(), tt.req)
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

func TestRemoveMaterial(t *testing.T) {
	tests := []struct {
		name          string
		materialID    string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:       "Success",
			materialID: "m1",
			setup: func(m *repository_test.Repopository) {
				m.DeleteMaterialFunc = func(ctx context.Context, materialID string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyMaterialID",
			materialID:    "",
			setup:         nil,
			expectedError: "material_id is required",
		},
		{
			name:       "NotFound",
			materialID: "m2",
			setup: func(m *repository_test.Repopository) {
				m.DeleteMaterialFunc = func(ctx context.Context, materialID string) error {
					return sql.ErrNoRows
				}
			},
			expectedError: "cannot delete: material id 'm2' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestRemoveMaterial case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.RemoveMaterial(context.Background(), tt.materialID)
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

func TestReorderNodes(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		req           model.ReorderNodesRequest
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			req:    model.ReorderNodesRequest{NodeIDs: []string{"n1", "n2"}},
			setup: func(m *repository_test.Repopository) {
				m.UpdateNodeSequenceFunc = func(ctx context.Context, nodeIDs []string) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyList",
			pathID:        "p1",
			req:           model.ReorderNodesRequest{NodeIDs: []string{}},
			setup:         nil,
			expectedError: "node_ids list cannot be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestReorderNodes case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			err := svc.ReorderNodes(context.Background(), tt.pathID, tt.req)
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

func TestGetNodeDetails(t *testing.T) {
	tests := []struct {
		name          string
		nodeID        string
		setup         func(*repository_test.Repopository)
		expectedError string
	}{
		{
			name:   "Success",
			nodeID: "n1",
			setup: func(m *repository_test.Repopository) {
				m.GetNodeByIDFunc = func(ctx context.Context, nodeID string) (*model.Node, error) {
					return &model.Node{NodeID: nodeID}, nil
				}
			},
			expectedError: "",
		},
		{
			name:          "EmptyNodeID",
			nodeID:        "",
			setup:         nil,
			expectedError: "node_id is required",
		},
		{
			name:   "NotFound",
			nodeID: "n2",
			setup: func(m *repository_test.Repopository) {
				m.GetNodeByIDFunc = func(ctx context.Context, nodeID string) (*model.Node, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: "node with id 'n2' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetNodeDetails case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			_, err := svc.GetNodeDetails(context.Background(), tt.nodeID)
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

func TestGetNodesByPathID(t *testing.T) {
	tests := []struct {
		name          string
		pathID        string
		setup         func(*repository_test.Repopository)
		expectedCount int
		expectedError string
	}{
		{
			name:   "Success",
			pathID: "p1",
			setup: func(m *repository_test.Repopository) {
				m.GetNodesByPathIDFunc = func(ctx context.Context, pathID string) ([]model.Node, error) {
					return []model.Node{{NodeID: "n1"}, {NodeID: "n2"}}, nil
				}
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name:          "EmptyPathID",
			pathID:        "",
			setup:         nil,
			expectedError: "path_id is required",
		},
		{
			name:   "DatabaseError",
			pathID: "p2",
			setup: func(m *repository_test.Repopository) {
				m.GetNodesByPathIDFunc = func(ctx context.Context, pathID string) ([]model.Node, error) {
					return nil, apperror.NewInternal("db error")
				}
			},
			expectedCount: 0,
			expectedError: "internal server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\033[36mExecuting TestGetNodesByPathID case: %s\033[0m", tt.name)
			mock := &repository_test.Repopository{}
			if tt.setup != nil {
				tt.setup(mock)
			}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := service.NewService(mock, nil, logger)

			nodes, err := svc.GetNodesByPathID(context.Background(), tt.pathID)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(nodes) != tt.expectedCount {
					t.Errorf("Expected %d nodes, got %d", tt.expectedCount, len(nodes))
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			}
		})
	}
}
