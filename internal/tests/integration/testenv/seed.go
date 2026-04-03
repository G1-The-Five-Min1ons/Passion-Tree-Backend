package testenv

import (
	"context"
	"fmt"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
	"passiontree/internal/connection"

	"github.com/google/uuid"
)

// SeedUser is a test helper that creates a temporary valid User and Profile in the database.
// It returns the new userID and a cleanup function to immediately delete the user when the test finishes.
func SeedUser(db connection.Database) (string, func(), error) {
	return SeedUserWithRole(db, model.RoleStudent)
}

// SeedUserWithRole creates a temporary valid User and Profile with the requested role.
// It returns the new userID and a cleanup function to immediately delete the user when the test finishes.
func SeedUserWithRole(db connection.Database, role model.UserRole) (string, func(), error) {
	repo := repository.NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Generate unique mock data
	mockID := uuid.New().String()
	user := &model.User{
		UserID:          mockID,
		Username:        "integration_tester_" + mockID[:8],
		Email:           fmt.Sprintf("integration_%s@example.com", mockID[:8]),
		Password:        "hashed_test_password",
		FirstName:       "Integration",
		LastName:        "Tester",
		Role:            role,
		HeartCount:      0,
		IsEmailVerified: true,
		AuthProvider:    model.AuthProviderLocal,
	}

	profile := &model.Profile{
		ProfileID:      uuid.New().String(),
		AvatarURL:      "http://example.com/avatar.png",
		RankName:       "Novice",
		LearningStreak: 0,
		LearningCount:  0,
		Location:       "Test Server",
		Bio:            "I am an automated integration test user.",
		UserID:         mockID,
	}

	// 2. Insert the user into the database
	userID, err := repo.CreateUser(ctx, user, profile)
	if err != nil {
		return "", nil, fmt.Errorf("failed to seed test user: %w", err)
	}

	// 3. Define the cleanup procedure
	cleanup := func() {
		delCtx, delCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer delCancel()
		_ = repo.DeleteUser(delCtx, userID)
	}

	return userID, cleanup, nil
}

func SeedRecommendationData(db connection.Database, userID string) (string, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sqlDB := db.GetDB()

	pathID := uuid.New().String()
	treeID := uuid.New().String()
	treeNodeID := uuid.New().String()
	reflectID := uuid.New().String()
	albumID := uuid.New().String()
	enrollID := uuid.New().String()

	_, err := sqlDB.ExecContext(ctx, `
		INSERT INTO dbo.Learning_Path (path_id, title, description, creator_id, publish_status, avg_rating) 
		VALUES (@p1, 'Test Path', 'Test Description', @p2, 'published', 4.5)`,
		pathID, userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert mock path: %w", err)
	}

	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO dbo.Path_Enroll (enroll_id, user_id, path_id) 
		VALUES (@p1, @p2, @p3)`,
		enrollID, userID, pathID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert mock path_enroll: %w", err)
	}

	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO dbo.Tree_Album (album_id, user_id) 
		VALUES (@p1, @p2)`,
		albumID, userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert mock tree_album: %w", err)
	}

	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO dbo.Tree (tree_id, path_id, album_id, node_count) 
		VALUES (@p1, @p2, @p3, 1)`,
		treeID, pathID, albumID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert mock tree: %w", err)
	}

	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO dbo.Tree_Node (tree_node_id, tree_id) 
		VALUES (@p1, @p2)`,
		treeNodeID, treeID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert mock tree_node: %w", err)
	}

	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO dbo.Reflect (reflect_id, tree_node_id, summary, primary_emotion, struggle_point, weighted_reflection_score, create_at) 
		VALUES (@p1, @p2, 'This is a test summary', 'Happy', 'None', 8.5, GETDATE())`,
		reflectID, treeNodeID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert mock reflect: %w", err)
	}

	cleanup := func() {
		delCtx, delCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer delCancel()

		_, _ = sqlDB.ExecContext(delCtx, `DELETE FROM dbo.Reflect WHERE reflect_id = @p1`, reflectID)
		_, _ = sqlDB.ExecContext(delCtx, `DELETE FROM dbo.Tree_Node WHERE tree_node_id = @p1`, treeNodeID)
		_, _ = sqlDB.ExecContext(delCtx, `DELETE FROM dbo.Tree WHERE tree_id = @p1`, treeID)
		_, _ = sqlDB.ExecContext(delCtx, `DELETE FROM dbo.Tree_Album WHERE album_id = @p1`, albumID)
		_, _ = sqlDB.ExecContext(delCtx, `DELETE FROM dbo.Path_Enroll WHERE enroll_id = @p1`, enrollID)
		_, _ = sqlDB.ExecContext(delCtx, `DELETE FROM dbo.Learning_Path WHERE path_id = @p1`, pathID)
	}

	return treeID, cleanup, nil
}
