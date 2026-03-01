package recommendation_test

import (
	"context"
	"testing"
	"time"

	"passiontree/internal/recommendation/repository"
	"passiontree/internal/tests/integration/testenv"
)

func TestRecommendation_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	db := testenv.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	reflections, currentPathID, err := repo.GetUserReflectionsByTree(ctx, testUserID, testTreeID)

	if err != nil {
		t.Fatalf("Failed to execute GetUserReflectionsByTree: %v", err)
	}

	t.Logf("Successfully executed query. Found %d reflections. Current Path ID: %s", len(reflections), currentPathID)

	if len(reflections) == 0 {
		t.Errorf("Expected at least 1 reflection, but got 0")
	} else {
		if reflections[0].Summary != "This is a test summary" {
			t.Errorf("Expected summary to be 'This is a test summary', got '%s'", reflections[0].Summary)
		}
	}
}
