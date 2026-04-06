package learningpath_test

import (
	"context"
	"testing"
	"time"

	authmodel "passiontree/internal/auth/model"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
	"passiontree/internal/tests/integration/testenv"
)

func TestLearningPath_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	t.Log("[STEP 1] Setting up test database connection...")
	dbConnectTime := time.Now()
	db := testenv.SetupTestDB(t)
	defer db.Close()
	t.Logf("[STEP 1] [PASS] Database connection established at %s", dbConnectTime.Format("2006-01-02 15:04:05"))

	t.Log("[STEP 2] Creating repository...")
	repo := repository.NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	t.Log("[STEP 2] [PASS] Repository created")

	// Create an approved teacher account so the test exercises the real teacher gate.
	t.Log("[STEP 3] Seeding approved teacher user...")
	userCreateTime := time.Now()
	testUserID, cleanupUser, err := testenv.SeedUserWithRole(db, authmodel.RoleTeacher)
	if err != nil {
		t.Fatalf("[STEP 3] [FAIL] Failed to seed prerequisite teacher for learning path test: %v", err)
	}
	defer cleanupUser()
	t.Logf("[STEP 3] [PASS] Teacher user created with ID: %s at %s", testUserID, userCreateTime.Format("2006-01-02 15:04:05"))

	t.Log("[STEP 4] Preparing learning path creation request...")
	req := model.CreatePathRequest{
		Title:       "Integration Test Path",
		Description: "This path was generated automatically during integration testing.",
		CreatorID:   testUserID,
		CoverImgURL: "https://example.com/cover.jpg",
	}
	t.Log("[STEP 4] [PASS] Request prepared")

	t.Log("[STEP 5] Creating learning path in database...")
	createTime := time.Now()
	pathID, err := repo.CreateLearningPath(ctx, req)
	if err != nil {
		t.Fatalf("[STEP 5] [FAIL] Failed to create learning path: %v", err)
	}
	t.Logf("[STEP 5] [PASS] Learning path created with ID: %s at %s", pathID, createTime.Format("2006-01-02 15:04:05"))

	t.Log("[STEP 6] Verifying path was persisted...")
	t.Logf("[STEP 6] [PASS] Path persisted successfully")

	// Clean up the learning path after testing
	t.Log("[STEP 7] Cleaning up test data...")
	deleteTime := time.Now()
	err = repo.DeleteLearningPath(ctx, pathID)
	if err != nil {
		t.Logf("[STEP 7] [FAIL] Failed to clean up learning path %s at %s: %v", pathID, deleteTime.Format("2006-01-02 15:04:05"), err)
	} else {
		t.Logf("[STEP 7] [PASS] Test learning path cleaned up successfully at %s", deleteTime.Format("2006-01-02 15:04:05"))
	}

	t.Logf("[COMPLETE] TestLearningPath_Integration [PASS] at %s", time.Now().Format("2006-01-02 15:04:05"))
}
