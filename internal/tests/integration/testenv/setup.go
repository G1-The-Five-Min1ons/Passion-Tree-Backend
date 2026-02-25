package testenv

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"passiontree/internal/config"
	"passiontree/internal/connection"
)

// SetupTestDB connects to the real test database defined in the root .env file.
func SetupTestDB(t *testing.T) connection.Database {
	t.Helper()

	// Navigating up to the root folder to find the .env file.
	// Since tests run from subdirectories of /internal/tests/integration/ we are 4 levels deep:
	// /internal/tests/integration/<domain> -> Need to go up 4 levels to find /.env
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Try loading from Backend root first, then Infrastructure root
	backendEnv := filepath.Join(wd, "..", "..", "..", "..", ".env")
	infraEnv := filepath.Join(wd, "..", "..", "..", "..", "..", "Passion-Tree-Infrastructure", ".env")

	if err := godotenv.Load(backendEnv); err != nil {
		if err2 := godotenv.Load(infraEnv); err2 != nil {
			t.Logf("Note: godotenv failed to load .env from Backend and Infrastructure folders: %v, %v", err, err2)
		}
	}

	// Now that environment variables are loaded, we can use config.LoadDBConfig
	cfg, err := config.LoadDBConfig()
	if err != nil {
		t.Fatalf("Failed to load local config: %v", err)
	}

	// We only need a short retry delay for local test booting
	db, err := connection.NewDatabaseWithRetry(cfg.DBConnString, 3, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to connect to integration test database via .env file: %v", err)
	}

	return db
}
