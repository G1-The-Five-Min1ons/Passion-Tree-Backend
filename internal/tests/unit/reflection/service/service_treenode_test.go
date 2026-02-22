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

func TestCreateTreeNode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req := model.CreateTreeNodeRequest{
			NodeTitle: "Day 1",
			NodeID:    "n1",
			TreeID:    "tree-1",
		}

		mock := &repository_test.MockRefRepo{
			AddSingleTreeNodeFunc: func(ctx context.Context, req model.CreateTreeNodeRequest) (string, error) {
				return "node-1", nil
			},
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{TreeID: treeID}, nil
			},
			GetTreeNodeByIDFunc: func(ctx context.Context, nodeID string) (*model.TreeNode, error) {
				return &model.TreeNode{TreeNodeID: nodeID}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		resp, err := svc.CreateTreeNode(context.Background(), req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if resp == nil || resp.TreeNodeID != "node-1" {
			t.Errorf("Expected valid TreeNodeResponse, got %v", resp)
		}
	})

	t.Run("MissingNodeTitle", func(t *testing.T) {
		req := model.CreateTreeNodeRequest{TreeID: "tree-1"}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.MockRefRepo{}, nil, logger)

		_, err := svc.CreateTreeNode(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "node_title is required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("TreeNotFound", func(t *testing.T) {
		req := model.CreateTreeNodeRequest{NodeTitle: "Day 1", NodeID: "n1", TreeID: "tree-2"}
		mock := &repository_test.MockRefRepo{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return nil, sql.ErrNoRows
			},
			AddSingleTreeNodeFunc: func(ctx context.Context, req model.CreateTreeNodeRequest) (string, error) {
				return "", sql.ErrNoRows
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.CreateTreeNode(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "invalid tree_id") && !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected invalid tree error, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		req := model.CreateTreeNodeRequest{NodeTitle: "Day 1", NodeID: "n1", TreeID: "tree-1"}
		mock := &repository_test.MockRefRepo{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{TreeID: treeID}, nil
			},
			AddSingleTreeNodeFunc: func(ctx context.Context, req model.CreateTreeNodeRequest) (string, error) {
				return "", errors.New("db limit")
			},
			GetTreeNodeByIDFunc: func(ctx context.Context, nodeID string) (*model.TreeNode, error) {
				return nil, sql.ErrNoRows
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.CreateTreeNode(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal server error, got %v", err)
		}
	})
}

func TestGetTreeNodesByTreeID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			GetTreeNodesByTreeIDFunc: func(ctx context.Context, treeID string) ([]model.TreeNode, error) {
				return []model.TreeNode{{TreeNodeID: "node1"}}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		nodes, err := svc.GetTreeNodesByTreeID(context.Background(), "t1")
		if err != nil || len(nodes) != 1 || nodes[0].TreeNodeID != "node1" {
			t.Errorf("Failed to successfully get trees: err=%v, nodes=%v", err, nodes)
		}
	})

	t.Run("MissingTreeID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.MockRefRepo{}, nil, logger)

		_, err := svc.GetTreeNodesByTreeID(context.Background(), "")
		if err == nil || !strings.Contains(err.Error(), "tree_id is required") {
			t.Errorf("Expected tree_id validation error, got %v", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			GetTreeNodesByTreeIDFunc: func(ctx context.Context, treeID string) ([]model.TreeNode, error) {
				return nil, errors.New("db limit")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetTreeNodesByTreeID(context.Background(), "t2")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal server error, got %v", err)
		}
	})
}

func TestGetTreeNodeByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			GetTreeNodeByIDFunc: func(ctx context.Context, nodeID string) (*model.TreeNode, error) {
				return &model.TreeNode{TreeNodeID: nodeID}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		node, err := svc.GetTreeNodeByID(context.Background(), "n1")
		if err != nil || node == nil || node.TreeNodeID != "n1" {
			t.Errorf("Failed to successfully get node: err=%v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			GetTreeNodeByIDFunc: func(ctx context.Context, nodeID string) (*model.TreeNode, error) {
				return nil, sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetTreeNodeByID(context.Background(), "n2")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})
}

func TestUpdateTreeNode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			UpdateTreeNodeFunc: func(ctx context.Context, nodeID string, req model.UpdateTreeNodeRequest) error {
				return nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateTreeNodeRequest{NodeTitle: "Edited"}
		err := svc.UpdateTreeNode(context.Background(), "n1", req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			UpdateTreeNodeFunc: func(ctx context.Context, nodeID string, req model.UpdateTreeNodeRequest) error {
				return sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateTreeNodeRequest{NodeTitle: "Edited"}
		err := svc.UpdateTreeNode(context.Background(), "n2", req)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.MockRefRepo{}, nil, logger)
		err := svc.UpdateTreeNode(context.Background(), "n2", model.UpdateTreeNodeRequest{})
		if err == nil || !strings.Contains(err.Error(), "node_title is required") {
			t.Errorf("Expected empty body validation error")
		}
	})
}

func TestDeleteTreeNode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			DeleteTreeNodeFunc: func(ctx context.Context, nodeID string) error {
				return nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		err := svc.DeleteTreeNode(context.Background(), "n1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.MockRefRepo{
			DeleteTreeNodeFunc: func(ctx context.Context, nodeID string) error {
				return sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		err := svc.DeleteTreeNode(context.Background(), "n2")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})
}
