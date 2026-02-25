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
		Role:            model.RoleStudent,
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
