package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

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

	t.Run("MissingTitle", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "", Difficulties: "Easy", PathID: "p1", AlbumID: "a1"}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "title is required") {
			t.Errorf("Expected title is required error")
		}
	})

	t.Run("MissingDifficulties", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "T", Difficulties: "", PathID: "p1", AlbumID: "a1"}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "difficulties is required") {
			t.Errorf("Expected difficulties is required error")
		}
	})

	t.Run("MissingPathID", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "T", Difficulties: "Easy", PathID: "", AlbumID: "a1"}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "path_id is required") {
			t.Errorf("Expected validation error")
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

	t.Run("CreateTreeForeignKeyError", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "T", Difficulties: "Easy", PathID: "p1", AlbumID: "album-2"}
		mock := &repository_test.Repository{
			CreateTreeFunc: func(ctx context.Context, req model.CreateTreeRequest) (string, error) {
				return "", errors.New("foreign key constraint error")
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "invalid album_id or path_id") {
			t.Errorf("Expected invalid album or path error, got %v", err)
		}
	})

	t.Run("GetTreeByIDFailure", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "T", Difficulties: "E", PathID: "p1", AlbumID: "a1"}
		mock := &repository_test.Repository{
			CreateTreeFunc: func(ctx context.Context, req model.CreateTreeRequest) (string, error) {
				return "tr1", nil
			},
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return nil, errors.New("db limit")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected db error, got %v", err)
		}
	})

	t.Run("GetTreeNodesByTreeIDFailure", func(t *testing.T) {
		req := model.CreateTreeRequest{Title: "T", Difficulties: "E", PathID: "p1", AlbumID: "a1"}
		mock := &repository_test.Repository{
			CreateTreeFunc: func(ctx context.Context, req model.CreateTreeRequest) (string, error) {
				return "tr1", nil
			},
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{}, nil
			},
			GetTreeNodesByTreeIDFunc: func(ctx context.Context, treeID string) ([]model.TreeNode, error) {
				return nil, errors.New("nodes fail")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.CreateTree(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected nodes failure error, got %v", err)
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

	t.Run("SyncsComputedStatusToRepository", func(t *testing.T) {
		lastReflectAt := time.Now().Add(-31 * 24 * time.Hour)
		var syncedTreeID, syncedStatus string

		mock := &repository_test.Repository{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{
					TreeID:        treeID,
					Difficulties:  "easy",
					Status:        "growing",
					LastReflectAt: &lastReflectAt,
				}, nil
			},
			UpdateTreeStatusFunc: func(ctx context.Context, treeID string, status string) error {
				syncedTreeID = treeID
				syncedStatus = status
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		tree, err := svc.GetTreeByID(context.Background(), "t-sync")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if tree.Status != "fading" {
			t.Fatalf("Expected computed status fading, got %s", tree.Status)
		}
		if syncedTreeID != "t-sync" || syncedStatus != "fading" {
			t.Fatalf("Expected status sync for tree t-sync => fading, got tree=%s status=%s", syncedTreeID, syncedStatus)
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

	t.Run("EmptyTreeID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)

		_, err := svc.GetTreeByID(context.Background(), "")
		if err == nil || !strings.Contains(err.Error(), "tree_id is required") {
			t.Errorf("Expected tree_id validation error")
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return nil, errors.New("db internal")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetTreeByID(context.Background(), "t2")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error, got %v", err)
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

		res, err := svc.GetTreesByAlbumID(context.Background(), "a1", false, "user1")
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
			GetTreesWithNodesByAlbumIDFunc: func(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error) {
				return []model.TreeResponse{{TreeID: "t2"}}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetTreesByAlbumID(context.Background(), "a2", true, "user1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		responses, ok := res.([]model.TreeResponse)
		if !ok || len(responses) != 1 {
			t.Errorf("Expected slice of 1 treeresponse, got %v", res)
		}
	})

	t.Run("WithNodesSyncsComputedStatusUsingPausedAt", func(t *testing.T) {
		lastReflectAt := time.Now().Add(-10 * 24 * time.Hour)
		pausedAt := time.Now().Add(-3 * 24 * time.Hour)
		var syncedTreeID, syncedStatus string

		mock := &repository_test.Repository{
			GetTreesWithNodesByAlbumIDFunc: func(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error) {
				return []model.TreeResponse{{
					TreeID:        "t-paused",
					Difficulties:  "medium",
					Status:        "growing",
					IsPause:       true,
					LastReflectAt: &lastReflectAt,
					PausedAt:      &pausedAt,
				}}, nil
			},
			UpdateTreeStatusFunc: func(ctx context.Context, treeID string, status string) error {
				syncedTreeID = treeID
				syncedStatus = status
				return nil
			},
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetTreesByAlbumID(context.Background(), "album-1", true, "user-1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		responses, ok := res.([]model.TreeResponse)
		if !ok || len(responses) != 1 {
			t.Fatalf("Expected one tree response, got %v", res)
		}
		if responses[0].Status != "fading" {
			t.Fatalf("Expected paused tree status fading, got %s", responses[0].Status)
		}
		if syncedTreeID != "t-paused" || syncedStatus != "fading" {
			t.Fatalf("Expected paused tree sync for t-paused => fading, got tree=%s status=%s", syncedTreeID, syncedStatus)
		}
	})

	t.Run("MissingAlbumID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)
		_, err := svc.GetTreesByAlbumID(context.Background(), "", false, "user1")
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

		_, err := svc.GetTreesByAlbumID(context.Background(), "a3", false, "user1")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal server error, got %v", err)
		}
	})

	t.Run("WithoutNodesEmptyList", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreesByAlbumIDFunc: func(ctx context.Context, albumID string) ([]model.Tree, error) {
				return []model.Tree{}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetTreesByAlbumID(context.Background(), "a1", false, "user1")
		if err != nil {
			t.Fatalf("Expected no error for empty tree list, got %v", err)
		}

		trees, ok := res.([]model.Tree)
		if !ok || len(trees) != 0 {
			t.Fatalf("Expected empty []model.Tree, got %v", res)
		}
	})

	t.Run("WithNodesEmptyList", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreesWithNodesByAlbumIDFunc: func(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error) {
				return []model.TreeResponse{}, nil
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetTreesByAlbumID(context.Background(), "empty", true, "user1")
		if err != nil {
			t.Fatalf("Expected no error for empty tree list with nodes, got %v", err)
		}

		responses, ok := res.([]model.TreeResponse)
		if !ok || len(responses) != 0 {
			t.Fatalf("Expected empty []model.TreeResponse, got %v", res)
		}
	})

	t.Run("WithNodesNotFound", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreesWithNodesByAlbumIDFunc: func(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error) {
				return nil, sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		res, err := svc.GetTreesByAlbumID(context.Background(), "a-unknown", true, "user1")
		if err != nil {
			t.Fatalf("Expected no error for sql.ErrNoRows with nodes, got %v", err)
		}

		responses, ok := res.([]model.TreeResponse)
		if !ok || len(responses) != 0 {
			t.Fatalf("Expected empty []model.TreeResponse on not found, got %v", res)
		}
	})

	t.Run("WithNodesDatabaseError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreesWithNodesByAlbumIDFunc: func(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error) {
				return nil, errors.New("db failed")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		_, err := svc.GetTreesByAlbumID(context.Background(), "err", true, "user1")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal server error with nodes")
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

	t.Run("EmptyTreeID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)
		err := svc.UpdateTree(context.Background(), "", model.UpdateTreeRequest{Title: "titler"})
		if err == nil || !strings.Contains(err.Error(), "tree_id is required") {
			t.Errorf("Expected validation error")
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		mock := &repository_test.Repository{
			UpdateTreeFunc: func(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
				return errors.New("db error")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		err := svc.UpdateTree(context.Background(), "t2", model.UpdateTreeRequest{Title: "e"})
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal error")
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

	t.Run("EmptyTreeID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)

		err := svc.DeleteTree(context.Background(), "")
		if err == nil || !strings.Contains(err.Error(), "tree_id is required") {
			t.Errorf("Expected not found error")
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		mock := &repository_test.Repository{
			DeleteTreeFunc: func(ctx context.Context, treeID string) error {
				return errors.New("test intern")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		err := svc.DeleteTree(context.Background(), "t2")
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected internal err")
		}
	})
}

func TestPauseTree(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		getTreeCalls := 0
		lastReflectAt := time.Now().Add(-31 * 24 * time.Hour)
		var syncedStatus string
		mock := &repository_test.Repository{
			PauseTreeFunc: func(ctx context.Context, treeID string, isPause bool) error {
				return nil
			},
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				getTreeCalls++
				if getTreeCalls == 1 {
					return &model.Tree{TreeID: treeID, IsPause: false, Difficulties: "easy", Status: "growing", LastReflectAt: &lastReflectAt}, nil
				}
				pausedAt := time.Now()
				return &model.Tree{TreeID: treeID, IsPause: true, Difficulties: "easy", Status: "growing", LastReflectAt: &lastReflectAt, PausedAt: &pausedAt}, nil
			},
			UpdateTreeStatusFunc: func(ctx context.Context, treeID string, status string) error {
				syncedStatus = status
				return nil
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
		if syncedStatus != "fading" {
			t.Errorf("Expected paused tree status to sync as fading, got %s", syncedStatus)
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

	t.Run("EmptyTreeID", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(&repository_test.Repository{}, nil, logger)
		_, err := svc.PauseTree(context.Background(), "", model.PauseTreeRequest{})
		if err == nil || !strings.Contains(err.Error(), "tree_id is required") {
			t.Errorf("Expected validation error")
		}
	})

	t.Run("GetTreeInternalError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return nil, errors.New("fake db")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		isPause := true
		req := model.PauseTreeRequest{IsPause: &isPause}
		_, err := svc.PauseTree(context.Background(), "t2", req)
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected fake db error")
		}
	})

	t.Run("PauseUpdateInternalError", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{TreeID: treeID}, nil
			},
			PauseTreeFunc: func(ctx context.Context, treeID string, isPause bool) error {
				return errors.New("update err")
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		isPause := false
		req := model.PauseTreeRequest{IsPause: &isPause}
		_, err := svc.PauseTree(context.Background(), "t2", req)
		if err == nil || !strings.Contains(err.Error(), "internal server error") {
			t.Errorf("Expected update error")
		}
	})

	t.Run("PauseUpdateNotFoundErr", func(t *testing.T) {
		mock := &repository_test.Repository{
			GetTreeByIDFunc: func(ctx context.Context, treeID string) (*model.Tree, error) {
				return &model.Tree{TreeID: treeID}, nil
			},
			PauseTreeFunc: func(ctx context.Context, treeID string, isPause bool) error {
				return sql.ErrNoRows
			},
		}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		svc := service.NewService(mock, nil, logger)

		isPause := false
		req := model.PauseTreeRequest{IsPause: &isPause}
		_, err := svc.PauseTree(context.Background(), "t2", req)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error")
		}
	})
}
