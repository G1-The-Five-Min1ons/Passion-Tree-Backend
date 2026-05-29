package reflection_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
	"passiontree/internal/reflection/service"
	"passiontree/internal/tests/integration/testenv"

	"github.com/google/uuid"
)

// TestTreeStatusSyncOnGetTreeByID_Integration verifies stale status values are
// synchronized back to tree.status in the real database during read flow.
func TestTreeStatusSyncOnGetTreeByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	db := testenv.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewService(repo, nil, nil, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID, cleanupUser, err := testenv.SeedUser(db)
	if err != nil {
		t.Fatalf("Failed to seed test user: %v", err)
	}
	defer cleanupUser()

	albumID, err := repo.CreateAlbum(ctx, model.CreateAlbumRequest{
		UserID:    userID,
		AlbumName: "Tree Status Sync Integration Album",
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}
	defer func() {
		_ = repo.DeleteAlbum(context.Background(), albumID)
	}()

	pathID := uuid.New().String()
	_, err = db.GetDB().ExecContext(ctx, `
		INSERT INTO learning_path (path_id, title, description, creator_id, publish_status, avg_rating)
		VALUES (@p1, 'Status Sync Path', 'Test path for reflection sync', @p2, 'published', 4.0)
	`, pathID, userID)
	if err != nil {
		t.Fatalf("Failed to insert test learning path: %v", err)
	}
	defer func() {
		_, _ = db.GetDB().ExecContext(context.Background(), `DELETE FROM learning_path WHERE path_id = @p1`, pathID)
	}()

	treeID := uuid.New().String()
	staleLastReflectAt := time.Now().Add(-31 * 24 * time.Hour) // easy => fading window

	_, err = db.GetDB().ExecContext(ctx, `
		INSERT INTO tree (
			tree_id, title, difficulties, path_id, status, is_pause, node_count,
			create_at, last_update, album_id, last_reflect_at, paused_at
		)
		VALUES (
			@p1, 'Status Sync Tree', 'easy', @p2, 'growing', 0, 0,
			GETDATE(), GETDATE(), @p3, @p4, NULL
		)
	`, treeID, pathID, albumID, staleLastReflectAt)
	if err != nil {
		t.Fatalf("Failed to insert test tree: %v", err)
	}
	defer func() {
		_, _ = db.GetDB().ExecContext(context.Background(), `DELETE FROM tree WHERE tree_id = @p1`, treeID)
	}()

	tree, err := svc.GetTreeByID(ctx, treeID)
	if err != nil {
		t.Fatalf("GetTreeByID failed: %v", err)
	}
	if tree.Status != "fading" {
		t.Fatalf("Expected computed status fading, got %s", tree.Status)
	}

	var persistedStatus string
	err = db.GetDB().QueryRowContext(ctx, `
		SELECT LOWER(LTRIM(RTRIM(status)))
		FROM tree
		WHERE tree_id = @p1
	`, treeID).Scan(&persistedStatus)
	if err != nil {
		t.Fatalf("Failed to read persisted tree status: %v", err)
	}

	if persistedStatus != "fading" {
		t.Fatalf("Expected persisted status fading, got %s", persistedStatus)
	}

	t.Logf("Status sync verified for tree %s: persisted=%s", treeID, persistedStatus)
}
