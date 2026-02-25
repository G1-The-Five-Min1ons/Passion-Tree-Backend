package learningpath_test

import (
	"context"
	"testing"
	"time"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
	"passiontree/internal/tests/integration/testenv"
)

func TestLearningPath_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	db := testenv.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Note: Create a dynamic test user to satisfy the DB foreign key constraint.
	testUserID, cleanupUser, err := testenv.SeedUser(db)
	if err != nil {
		t.Fatalf("Failed to seed prerequisite user for learning path test: %v", err)
	}
	defer cleanupUser()

	req := model.CreatePathRequest{
		Title:       "Integration Test Path",
		Description: "This path was generated automatically during integration testing.",
		CreatorID:   testUserID,
		CoverImgURL: "https://example.com/cover.jpg",
	}

	pathID, err := repo.CreateLearningPath(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create learning path: %v", err)
	}

	t.Logf("Successfully inserted learning path with ID: %s", pathID)

	// Clean up the learning path after testing
	err = repo.DeleteLearningPath(ctx, pathID)
	if err != nil {
		t.Logf("Warning: Failed to clean up learning path %s: %v", pathID, err)
	} else {
		t.Logf("Successfully cleaned up test learning path from the database")
	}
}
