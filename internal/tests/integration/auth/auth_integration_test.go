package auth_test

import (
	"context"
	"testing"
	"time"

	"passiontree/internal/auth/repository"
	"passiontree/internal/tests/integration/testenv"
)

func TestAuth_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	db := testenv.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Auth integration test: Verify that fetching a non-existent user behaves correctly
	// rather than crashing the SQL driver.
	randomEmail := "integration_non_existent_user@example.com"
	user, err := repo.GetUserByEmail(ctx, randomEmail)

	if err != nil {
		t.Fatalf("Expected no error for querying a non-existent email, got: %v", err)
	}

	if user != nil {
		t.Fatalf("Expected nil user for non-existent email, but found one")
	}

	t.Logf("Successfully verified Auth database queries using real connection.")
}
