package learningpath_test

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"passiontree/internal/auth/model"
	learningmodel "passiontree/internal/learning-path/model"
	learningrepo "passiontree/internal/learning-path/repository"
	learningservice "passiontree/internal/learning-path/service"
	"passiontree/internal/tests/integration/testenv"
)

func TestLearningPath_DatabaseBackedFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	t.Log("[STEP 1] Setting up test database connection...")
	testStartTime := time.Now()
	db := testenv.SetupTestDB(t)
	defer db.Close()
	t.Logf("[STEP 1] [PASS] Database connection established at %s", testStartTime.Format("2006-01-02 15:04:05"))

	// Diagnostic: Check which database we're connected to
	t.Log("[STEP 1.5] Verifying database connection details...")
	var dbName, serverInfo string
	err := db.GetDB().QueryRowContext(context.Background(), `SELECT DB_NAME(), @@SERVERNAME`).Scan(&dbName, &serverInfo)
	if err == nil {
		t.Logf("[STEP 1.5] [INFO] Connected to database: %s on server: %s", dbName, serverInfo)
	} else {
		t.Logf("[STEP 1.5] [WARN] Could not get database info: %v", err)
	}

	t.Log("[STEP 2] Creating context and initializing services...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pathRepo := learningrepo.NewRepository(db)
	svc := learningservice.NewService(pathRepo, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Log("[STEP 2] [PASS] Services initialized")

	t.Log("[STEP 3] Seeding test users (teacher and student)...")
	userSeedTime := time.Now()
	teacherID, cleanupTeacher, err := testenv.SeedUserWithRole(db, model.RoleTeacher)
	if err != nil {
		t.Fatalf("[STEP 3] [FAIL] failed to seed teacher user: %v", err)
	}
	_ = cleanupTeacher
	t.Logf("[STEP 3] [PASS] Teacher user created with ID: %s at %s", teacherID, userSeedTime.Format("15:04:05"))

	studentID, cleanupStudent, err := testenv.SeedUserWithRole(db, model.RoleStudent)
	if err != nil {
		t.Fatalf("[STEP 3] [FAIL] failed to seed student user: %v", err)
	}
	_ = cleanupStudent
	t.Logf("[STEP 3] [PASS] Student user created with ID: %s at %s", studentID, userSeedTime.Format("15:04:05"))

	t.Log("[STEP 4] Setting environment variable for path creation...")
	t.Setenv("ALLOW_UNVERIFIED_TEACHER_PATH_CREATION", "true")
	t.Log("[STEP 4] [PASS] Environment configured")

	t.Log("[STEP 5] Creating learning path through service...")
	createPathTime := time.Now()
	pathID, err := svc.CreatePath(ctx, learningmodel.CreatePathRequest{
		Title:          "DB Integration Path",
		Objective:      "Prove teacher path creation writes to the database",
		Description:    "Created by the backend integration test.",
		CoverImgURL:    "https://example.com/cover.jpg",
		CreatorID:      teacherID,
		Publish_status: "draft",
	})
	if err != nil {
		t.Fatalf("[STEP 5] [FAIL] failed to create learning path through service: %v", err)
	}
	t.Logf("[STEP 5] [PASS] Learning path created with ID: %s at %s", pathID, createPathTime.Format("2006-01-02 15:04:05"))

	t.Log("[STEP 6] Creating 5 nodes for the learning path...")
	nodeCreationStartTime := time.Now()
	createdNodeIDs := make([]string, 0, 5)
	for index := 1; index <= 5; index++ {
		nodeCreationTime := time.Now()
		t.Logf("  [6.%d] Creating node %d at %s...", index, index, nodeCreationTime.Format("15:04:05"))
		nodeID, err := svc.AddNode(ctx, learningmodel.CreateNodeRequest{
			Title:       "DB Integration Node " + strconv.Itoa(index),
			Description: "Created by the backend integration test.",
			PathID:      pathID,
			Sequence:    strconv.Itoa(index),
			Link_vdo:    "https://example.com/node-" + strconv.Itoa(index),
		})
		if err != nil {
			t.Fatalf("[6.%d] [FAIL] failed to create node %d through service: %v", index, index, err)
		}
		createdNodeIDs = append(createdNodeIDs, nodeID)
		t.Logf("  [6.%d] [PASS] Node %d created with ID: %s at %s", index, index, nodeID, nodeCreationTime.Format("15:04:05"))
	}
	t.Logf("[STEP 6] [PASS] All 5 nodes created successfully (total time: %v)", time.Since(nodeCreationStartTime))

	// Verify database row count for nodes before proceeding
	t.Log("[STEP 6.5] Verifying nodes were persisted to database...")
	var nodeCount int
	err = db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM node WHERE path_id = @p1`, pathID).Scan(&nodeCount)
	if err != nil {
		t.Fatalf("[STEP 6.5] [FAIL] Failed to count nodes in database: %v", err)
	}
	if nodeCount != 5 {
		t.Fatalf("[STEP 6.5] [FAIL] Expected 5 nodes in database, but found %d. Nodes may not be persisting!", nodeCount)
	}
	t.Logf("[STEP 6.5] [PASS] ✓ Database verified: %d node rows exist in learning_path.node table", nodeCount)

	// Verify path was persisted to database
	var pathCount int
	err = db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_path WHERE path_id = @p1`, pathID).Scan(&pathCount)
	if err != nil {
		t.Fatalf("[STEP 6.5] [FAIL] Failed to count paths in database: %v", err)
	}
	if pathCount != 1 {
		t.Fatalf("[STEP 6.5] [FAIL] Expected 1 path in database, but found %d. Path may not be persisting!", pathCount)
	}
	t.Logf("[STEP 6.5] [PASS] ✓ Database verified: %d path row exists in learning_path.learning_path table", pathCount)

	// Force explicit connection flush to ensure all database writes are committed
	t.Log("[STEP 6.6] Flushing database to ensure all writes are committed...")
	const flushSQL = `SELECT @@TRANCOUNT`
	var tranCount int
	err = db.GetDB().QueryRowContext(ctx, flushSQL).Scan(&tranCount)
	if err == nil {
		t.Logf("[STEP 6.6] [INFO] Transaction count: %d", tranCount)
	}
	// Close all idle connections in pool and force reconnection to complete any pending operations
	db.GetDB().SetMaxOpenConns(0)
	time.Sleep(50 * time.Millisecond)
	db.GetDB().SetMaxOpenConns(10)
	time.Sleep(50 * time.Millisecond)
	t.Log("[STEP 6.6] [PASS] Connection pool flushed - pending writes should be committed")

	defer func() {
		cleanupTime := time.Now()
		t.Logf("[CLEANUP] Disabled at %s - Test data will be preserved in database for manual inspection", cleanupTime.Format("2006-01-02 15:04:05"))
		t.Logf("[CLEANUP] [INFO] You can inspect the following data in the database:")
		t.Logf("[CLEANUP]   - learning_path table: path_id=%s", pathID)
		t.Logf("[CLEANUP]   - node table: 5 nodes with path_id=%s", pathID)
		t.Logf("[CLEANUP]   - path_enroll table: enrollment for student_id=%s", studentID)
		t.Logf("[CLEANUP] [INFO] To delete test data manually, run:")
		t.Logf("[CLEANUP]   DELETE FROM node WHERE path_id = '%s'", pathID)
		t.Logf("[CLEANUP]   DELETE FROM path_enroll WHERE path_id = '%s'", pathID)
		t.Logf("[CLEANUP]   DELETE FROM learning_path WHERE path_id = '%s'", pathID)

		// Cleanup is DISABLED for debugging - comment out deletion queries below:
		// for index := len(createdNodeIDs) - 1; index >= 0; index-- {
		// 	_, _ = db.GetDB().ExecContext(context.Background(), `DELETE FROM node WHERE node_id = @p1`, createdNodeIDs[index])
		// }
		// _, _ = db.GetDB().ExecContext(context.Background(), `DELETE FROM path_enroll WHERE user_id = @p1 AND path_id = @p2`, studentID, pathID)
		// _, _ = db.GetDB().ExecContext(context.Background(), `DELETE FROM learning_path WHERE path_id = @p1`, pathID)
	}()

	t.Log("[STEP 7] Reading created path from database...")
	createdPath, err := pathRepo.GetLearningPathByID(ctx, pathID)
	if err != nil {
		t.Fatalf("[STEP 7] [FAIL] failed to read created learning path from database: %v", err)
	}
	if createdPath == nil {
		t.Fatalf("[STEP 7] [FAIL] created learning path was not persisted correctly")
	}
	if createdPath.Title != "DB Integration Path" {
		t.Fatalf("[STEP 7] [FAIL] expected persisted learning path title to match, got %q", createdPath.Title)
	}
	t.Logf("[STEP 7] [PASS] Path read successfully: %s", createdPath.Title)

	t.Log("[STEP 8] Updating learning path...")
	updatePathTime := time.Now()
	if err := svc.UpdatePath(ctx, pathID, learningmodel.UpdatePathRequest{
		Title:          "DB Integration Path Updated",
		Objective:      "Verify update writes back to the database",
		Description:    "Updated by the backend integration test.",
		CoverImgURL:    "https://example.com/cover-updated.jpg",
		Publish_status: "draft",
	}); err != nil {
		t.Fatalf("[STEP 8] [FAIL] failed to update learning path through service: %v", err)
	}
	t.Logf("[STEP 8] [PASS] Path updated successfully at %s", updatePathTime.Format("2006-01-02 15:04:05"))

	t.Log("[STEP 9] Verifying updated path in database...")
	updatedPath, err := pathRepo.GetLearningPathByID(ctx, pathID)
	if err != nil {
		t.Fatalf("[STEP 9] [FAIL] failed to read updated learning path from database: %v", err)
	}
	if updatedPath == nil {
		t.Fatalf("[STEP 9] [FAIL] updated learning path was not persisted correctly")
	}
	if updatedPath.Title != "DB Integration Path Updated" {
		t.Fatalf("[STEP 9] [FAIL] expected updated learning path title to match, got %q", updatedPath.Title)
	}
	if updatedPath.Objective != "Verify update writes back to the database" {
		t.Fatalf("[STEP 9] [FAIL] expected updated learning path objective to match, got %q", updatedPath.Objective)
	}
	if updatedPath.Description != "Updated by the backend integration test." {
		t.Fatalf("[STEP 9] [FAIL] expected updated learning path description to match, got %q", updatedPath.Description)
	}
	if updatedPath.CoverImgURL != "https://example.com/cover-updated.jpg" {
		t.Fatalf("[STEP 9] [FAIL] expected updated learning path cover image to match, got %q", updatedPath.CoverImgURL)
	}
	t.Log("[STEP 9] [PASS] Path update verified")

	// Verify database actually has the updated values
	t.Log("[STEP 9.5] Verifying updated data persisted in database...")
	var updatedTitle, updatedDesc string
	err = db.GetDB().QueryRowContext(ctx, `SELECT title, description FROM learning_path WHERE path_id = @p1`, pathID).Scan(&updatedTitle, &updatedDesc)
	if err != nil {
		t.Fatalf("[STEP 9.5] [FAIL] Failed to read updated path from database: %v", err)
	}
	if updatedTitle != "DB Integration Path Updated" || updatedDesc != "Updated by the backend integration test." {
		t.Fatalf("[STEP 9.5] [FAIL] Database values not updated: title=%q, desc=%q", updatedTitle, updatedDesc)
	}
	t.Logf("[STEP 9.5] [PASS] ✓ Database verified: Updated values persisted correctly - title=%q, description=%q", updatedTitle, updatedDesc)

	t.Log("[STEP 10] Retrieving all nodes for the path...")
	nodes, err := pathRepo.GetNodesByPathID(ctx, pathID, teacherID)
	if err != nil {
		t.Fatalf("[STEP 10] [FAIL] failed to fetch nodes for created learning path: %v", err)
	}
	if len(nodes) != 5 {
		t.Fatalf("[STEP 10] [FAIL] expected 5 persisted nodes, got %d", len(nodes))
	}
	t.Logf("[STEP 10] [PASS] All 5 nodes verified in database")

	t.Log("[STEP 11] Student enrolling in learning path...")
	enrollmentTime := time.Now()
	if err := svc.StartPath(ctx, pathID, studentID); err != nil {
		t.Fatalf("[STEP 11] [FAIL] failed to enroll student through service: %v", err)
	}
	t.Logf("[STEP 11] [PASS] Student successfully enrolled in path at %s", enrollmentTime.Format("2006-01-02 15:04:05"))

	// Verify enrollment was actually written to database
	t.Log("[STEP 11.5] Verifying enrollment data persisted in database...")
	var enrollmentStatus string
	err = db.GetDB().QueryRowContext(ctx, `SELECT enrollment_status FROM path_enroll WHERE path_id = @p1 AND user_id = @p2`, pathID, studentID).Scan(&enrollmentStatus)
	if err != nil {
		t.Fatalf("[STEP 11.5] [FAIL] Failed to read enrollment from database: %v", err)
	}
	if enrollmentStatus == "" {
		t.Fatalf("[STEP 11.5] [FAIL] Enrollment status empty in database!")
	}
	t.Logf("[STEP 11.5] [PASS] ✓ Database verified: Enrollment written to path_enroll table with status=%q", enrollmentStatus)

	t.Log("[STEP 12] Verifying enrollment status...")
	enrollment, err := svc.GetEnrollmentStatus(ctx, pathID, studentID)
	if err != nil {
		t.Fatalf("[STEP 12] [FAIL] failed to fetch enrollment status: %v", err)
	}
	if enrollment == nil || enrollment.Enrollment_status != "active" {
		t.Fatalf("[STEP 12] [FAIL] expected active enrollment in database, got %#v", enrollment)
	}
	t.Log("[STEP 12] [PASS] Enrollment status verified as active")

	t.Log("[STEP 13] Verifying path exists in database...")
	var pathCountFinal int
	if err := db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_path WHERE path_id = @p1`, pathID).Scan(&pathCountFinal); err != nil {
		t.Fatalf("[STEP 13] [FAIL] failed to verify learning_path row: %v", err)
	}
	if pathCountFinal != 1 {
		t.Fatalf("[STEP 13] [FAIL] expected 1 learning_path row, got %d", pathCountFinal)
	}
	t.Log("[STEP 13] [PASS] Path verified in database")

	t.Log("[STEP 14] Verifying enrollment record exists...")
	var enrollCount int
	if err := db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM path_enroll WHERE user_id = @p1 AND path_id = @p2`, studentID, pathID).Scan(&enrollCount); err != nil {
		t.Fatalf("[STEP 14] [FAIL] failed to verify path_enroll row: %v", err)
	}
	if enrollCount != 1 {
		t.Fatalf("[STEP 14] [FAIL] expected 1 path_enroll row, got %d", enrollCount)
	}
	t.Log("[STEP 14] [PASS] Enrollment record verified in database")

	t.Logf("[COMPLETE] TestLearningPath_DatabaseBackedFlow [PASS] - path_id=%s, student_id=%s at %s", pathID, studentID, time.Now().Format("2006-01-02 15:04:05"))
}
