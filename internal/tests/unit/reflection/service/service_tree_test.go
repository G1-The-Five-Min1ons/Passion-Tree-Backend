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

func TestCreateTree(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req := model.CreateTreeRequest{
			AlbumID:      "album-1",
			Title:        "New Tree",
			Difficulties: "Hard",
			PathID:       "path-1",
		}

		mock := &repository_test.Repository{
			CreateTreeFunc: func(ctx context.Context, req model.CreateTreeRequest) (string, error) {
				return "tree-1", nil
			},
			GetAlbumByIDFunc: func(ctx context.Context, albumID string) (*model.Album, error) {
				return &model.Album{AlbumID: albumID}, nil
			},
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{TreeID: treeID}, nil
			},
			GetTreeNodesByTreeIDFunc: func(ctx context.Context, treeID string) ([]model.TreeNode, error) {
				return []model.TreeNode{}, nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		resp, err := svc.CreateTree(context.Background(), req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if resp == nil || resp.TreeID != "tree-1" {
			t.Errorf("Expected valid TreeResponse, got %v", resp)
		}
	})

	t.Run("MissingAlbumID", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "T", Difficulties: "Easy", PathID: "p1"}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "album_id is required") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})

	t.Run("AlbumNotFound", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "T", Difficulties: "Easy", PathID: "p1", AlbumID: "album-2"}
		mock := &repository_test.Repository{
			GetAlbumByIDFunc: func(ctx context.Context, albumID string) (*model.Album, error) {
				return nil, sql.ErrNoRows
			},
			CreateTreeFunc: func(ctx context.Context, req model.CreateTreeRequest) (string, error) {
				return "", sql.ErrNoRows
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || (!strings.Contains(err.Error(), "invalid album_id") && !strings.Contains(err.Error(), "internal server error")) {
			t.Errorf("Expected invalid album error, got %v", err)
		}
	})
}

func TestGetTreeByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{TreeID: treeID}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		tree, err := svc.GetTreeByID(context.Background(), "t1")
		if err != nil || tree == nil || tree.TreeID != "t1" {
			t.Errorf("Failed to successfully get tree: err=%v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return nil, sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetTreeByID(context.Background(), "t2")
		if err == nil || !strings.Contains(err.Error(), "tree with id 't2' not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})
}

func TestGetTreesByAlbumID(t *testing.T) {
	t.Run("WithoutNodes", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreesByAlbumIDFunc: func(ctx context.Context, albumID string) ([]model.Tree, error) {
				return []model.Tree{{TreeID: "t1"}}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetTreesByAlbumID(context.Background(), "a1", false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		trees, ok := res.([]model.Tree)
		if !ok || len(trees) != 1 {
			t.Errorf("Expected slice of 1 tree, got %v", res)
		}
	})

	t.Run("WithNodes", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreesWithNodesByAlbumIDFunc: func(ctx context.Context, albumID string) ([]model.TreeResponse, error) {
				return []model.TreeResponse{{TreeID: "t2"}}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetTreesByAlbumID(context.Background(), "a2", true)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		responses, ok := res.([]model.TreeResponse)
		if !ok || len(responses) != 1 {
			t.Errorf("Expected slice of 1 treeresponse, got %v", res)
		}
	})

	t.Run("MissingAlbumID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)
		_, err := svc.GetTreesByAlbumID(context.Background(), "", false)
		if err == nil || !strings.Contains(err.Error(), "album_id is required") {
			t.Errorf("Expected validation error")
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreesByAlbumIDFunc: func(ctx context.Context, albumID string) ([]model.Tree, error) {
				return nil, errors.New("db limit")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetTreesByAlbumID(context.Background(), "a3", false)
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal server error, got %v", err)
		}
	})
}

func TestUpdateTree(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateTreeFunc: func(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
				return nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateTreeRequest{Title: "Edited"}
		err := svc.UpdateTree(context.Background(), "t1", req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateTreeFunc: func(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
				return sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		req := model.UpdateTreeRequest{Title: "Edited"}
		err := svc.UpdateTree(context.Background(), "t2", req)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)
		err := svc.UpdateTree(context.Background(), "t2", model.UpdateTreeRequest{})
		if err == nil || !strings.Contains(err.Error(), "title is required") {
			t.Errorf("Expected empty body validation error, got %v", err)
		}
	})
}

func TestDeleteTree(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			DeleteTreeFunc: func(ctx context.Context, treeID string) error {
				return nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		err := svc.DeleteTree(context.Background(), "t1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			DeleteTreeFunc: func(ctx context.Context, treeID string) error {
				return sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		err := svc.DeleteTree(context.Background(), "t2")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})
}

func TestPauseTree(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &repository_test.Repository{
			PauseTreeFunc: func(ctx context.Context, treeID string, isPause bool) error {
				return nil
			},
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{TreeID: treeID, IsPause: false}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		isPause := true
		req := model.PauseTreeRequest{IsPause: &isPause}
		res, err := svc.PauseTree(context.Background(), "t1", req)
		if err != nil || !res {
			t.Errorf("Expected successful pause, got res=%v, err=%v", res, err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			PauseTreeFunc: func(ctx context.Context, treeID string, isPause bool) error {
				return sql.ErrNoRows
			},
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return nil, sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		isPause := true
		req := model.PauseTreeRequest{IsPause: &isPause}
		_, err := svc.PauseTree(context.Background(), "t2", req)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got %v", err)
		}
	})
}
