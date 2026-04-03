package learningpath_test

import (
	"context"
	"io"
	"log/slog"
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

	db := testenv.SetupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pathRepo := learningrepo.NewRepository(db)
	svc := learningservice.NewService(pathRepo, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	teacherID, cleanupTeacher, err := testenv.SeedUserWithRole(db, model.RoleTeacher)
	if err != nil {
		t.Fatalf("failed to seed teacher user: %v", err)
	}
	defer cleanupTeacher()

	studentID, cleanupStudent, err := testenv.SeedUserWithRole(db, model.RoleStudent)
	if err != nil {
		t.Fatalf("failed to seed student user: %v", err)
	}
	defer cleanupStudent()

	t.Setenv("ALLOW_UNVERIFIED_TEACHER_PATH_CREATION", "true")

	pathID, err := svc.CreatePath(ctx, learningmodel.CreatePathRequest{
		Title:          "DB Integration Path",
		Objective:      "Prove teacher path creation writes to the database",
		Description:    "Created by the backend integration test.",
		CoverImgURL:    "https://example.com/cover.jpg",
		CreatorID:      teacherID,
		Publish_status: "draft",
	})
	if err != nil {
		t.Fatalf("failed to create learning path through service: %v", err)
	}
	defer func() {
		_, _ = db.GetDB().ExecContext(context.Background(), `DELETE FROM path_enroll WHERE user_id = @p1 AND path_id = @p2`, studentID, pathID)
		_, _ = db.GetDB().ExecContext(context.Background(), `DELETE FROM learning_path WHERE path_id = @p1`, pathID)
	}()

	createdPath, err := pathRepo.GetLearningPathByID(ctx, pathID)
	if err != nil {
		t.Fatalf("failed to read created learning path from database: %v", err)
	}
	if createdPath == nil {
		t.Fatalf("created learning path was not persisted correctly")
	}
	if createdPath.Title != "DB Integration Path" {
		t.Fatalf("expected persisted learning path title to match, got %q", createdPath.Title)
	}

	if err := svc.UpdatePath(ctx, pathID, learningmodel.UpdatePathRequest{
		Title:          "DB Integration Path Updated",
		Objective:      "Verify update writes back to the database",
		Description:    "Updated by the backend integration test.",
		CoverImgURL:    "https://example.com/cover-updated.jpg",
		Publish_status: "draft",
	}); err != nil {
		t.Fatalf("failed to update learning path through service: %v", err)
	}

	updatedPath, err := pathRepo.GetLearningPathByID(ctx, pathID)
	if err != nil {
		t.Fatalf("failed to read updated learning path from database: %v", err)
	}
	if updatedPath == nil {
		t.Fatalf("updated learning path was not persisted correctly")
	}
	if updatedPath.Title != "DB Integration Path Updated" {
		t.Fatalf("expected updated learning path title to match, got %q", updatedPath.Title)
	}
	if updatedPath.Objective != "Verify update writes back to the database" {
		t.Fatalf("expected updated learning path objective to match, got %q", updatedPath.Objective)
	}
	if updatedPath.Description != "Updated by the backend integration test." {
		t.Fatalf("expected updated learning path description to match, got %q", updatedPath.Description)
	}
	if updatedPath.CoverImgURL != "https://example.com/cover-updated.jpg" {
		t.Fatalf("expected updated learning path cover image to match, got %q", updatedPath.CoverImgURL)
	}

	if err := svc.StartPath(ctx, pathID, studentID); err != nil {
		t.Fatalf("failed to enroll student through service: %v", err)
	}

	enrollment, err := svc.GetEnrollmentStatus(ctx, pathID, studentID)
	if err != nil {
		t.Fatalf("failed to fetch enrollment status: %v", err)
	}
	if enrollment == nil || enrollment.Enrollment_status != "active" {
		t.Fatalf("expected active enrollment in database, got %#v", enrollment)
	}

	var pathCount int
	if err := db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_path WHERE path_id = @p1`, pathID).Scan(&pathCount); err != nil {
		t.Fatalf("failed to verify learning_path row: %v", err)
	}
	if pathCount != 1 {
		t.Fatalf("expected 1 learning_path row, got %d", pathCount)
	}

	var enrollCount int
	if err := db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM path_enroll WHERE user_id = @p1 AND path_id = @p2`, studentID, pathID).Scan(&enrollCount); err != nil {
		t.Fatalf("failed to verify path_enroll row: %v", err)
	}
	if enrollCount != 1 {
		t.Fatalf("expected 1 path_enroll row, got %d", enrollCount)
	}

	t.Logf("Database-backed learning path flow succeeded: path_id=%s student_id=%s", pathID, studentID)
}
