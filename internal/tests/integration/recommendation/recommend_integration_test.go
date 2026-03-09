package recommendation_test

import (
	"context"
	"testing"
	"time"

	"passiontree/internal/recommendation/repository"
	"passiontree/internal/tests/integration/testenv"
)

func TestRecommendationRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	db := testenv.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup: Seed User and Recommendation Mock Data
	testUserID, cleanupUser, err := testenv.SeedUser(db)
	if err != nil {
		t.Fatalf("Failed to seed prerequisite user: %v", err)
	}
	defer cleanupUser()

	testTreeID, cleanupData, err := testenv.SeedRecommendationData(db, testUserID)
	if err != nil {
		t.Fatalf("Failed to seed recommendation prerequisite data: %v", err)
	}
	defer cleanupData()

	t.Run("GetUserReflectionsByTree", func(t *testing.T) {
		reflections, currentPathID, err := repo.GetUserReflectionsByTree(ctx, testUserID, testTreeID)
		if err != nil {
			t.Fatalf("Failed to execute GetUserReflectionsByTree: %v", err)
		}

		t.Logf("Found %d reflections. Current Path ID: %s", len(reflections), currentPathID)

		if len(reflections) == 0 {
			t.Errorf("Expected at least 1 reflection, but got 0")
		} else {
			if reflections[0].Summary != "This is a test summary" {
				t.Errorf("Expected summary to be 'This is a test summary', got '%s'", reflections[0].Summary)
			}
		}
	})

	t.Run("GetUserEnrolledPathsForRec", func(t *testing.T) {
		paths, err := repo.GetUserEnrolledPathsForRec(ctx, testUserID)
		if err != nil {
			t.Fatalf("Failed to execute GetUserEnrolledPathsForRec: %v", err)
		}

		t.Logf("Found %d enrolled paths for user", len(paths))

		if len(paths) == 0 {
			t.Errorf("Expected at least 1 enrolled path, but got 0")
		} else {
			if paths[0].Title != "Test Path" {
				t.Errorf("Expected title to be 'Test Path', got '%s'", paths[0].Title)
			}
		}
	})

	t.Run("GetTopPopularPaths", func(t *testing.T) {
		popularPaths, err := repo.GetTopPopularPaths(ctx)
		if err != nil {
			t.Fatalf("Failed to execute GetTopPopularPaths: %v", err)
		}

		t.Logf("Found %d popular paths", len(popularPaths))

		if len(popularPaths) == 0 {
			t.Errorf("Expected to find popular paths, but got 0")
		} else if len(popularPaths) > 5 {
			t.Errorf("Expected maximum 5 popular paths, but got %d", len(popularPaths))
		}
	})
}
