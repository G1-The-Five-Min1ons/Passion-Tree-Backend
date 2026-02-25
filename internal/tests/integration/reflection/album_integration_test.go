package reflection_test

import (
	"context"
	"testing"
	"time"

	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
	"passiontree/internal/tests/integration/testenv"
)

// TestCreateAlbum_Integration verifies that we can connect to the real SQL Server
// database and execute an INSERT statement without any driver or schema errors.
func TestCreateAlbum_Integration(t *testing.T) {
	// Skip if we are in a short test run to avoid hitting the real DB
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	// 1. Connect to the real test database
	db := testenv.SetupTestDB(t)
	defer db.Close()

	// 2. Initialize the REAL repository
	repo := repository.NewRepository(db)

	// 3. Create a dynamic test user to satisfy the DB foreign key constraint.
	testUserID, cleanupUser, err := testenv.SeedUser(db)
	if err != nil {
		t.Fatalf("Failed to seed prerequisite user for album test: %v", err)
	}
	defer cleanupUser()

	req := model.CreateAlbumRequest{
		UserID:        testUserID,
		AlbumName:     "Integration Test Album",
		CoverImageURL: "https://example.com/cover.jpg",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 4. Make a real database call
	albumID, err := repo.CreateAlbum(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create album in real database: %v", err)
	}

	if albumID == "" {
		t.Fatalf("Expected a non-empty album ID to be returned")
	}

	t.Logf("Successfully inserted album with ID: %s", albumID)

	// 5. Cleanup the database so tests do not leave garbage behind
	err = repo.DeleteAlbum(ctx, albumID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup album %s: %v", albumID, err)
	} else {
		t.Logf("Successfully cleaned up test album from the database")
	}
}
