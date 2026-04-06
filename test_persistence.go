package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"passiontree/internal/config"
	"passiontree/internal/connection"

	"github.com/joho/godotenv"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	// Load .env from the backend or infrastructure location.
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	backendEnv := filepath.Join(wd, ".env")
	infraEnv := filepath.Join(wd, "..", "Passion-Tree-Infrastructure", ".env")
	if err := godotenv.Load(backendEnv); err != nil {
		if err2 := godotenv.Load(infraEnv); err2 != nil {
			fmt.Fprintf(os.Stderr, "Note: Could not load .env from backend or infrastructure: %v, %v\n", err, err2)
		}
	}

	// Load database config using the config package
	cfg, err := config.LoadDBConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load database config: %v\n", err)
		os.Exit(1)
	}

	// Create database connection
	db, err := connection.NewDatabase(cfg.DBConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	pathID := os.Getenv("PATH_ID")
	if len(os.Args) > 1 && os.Args[1] != "" {
		pathID = os.Args[1]
	}
	if pathID == "" {
		pathID = "b90c1c5f-7687-4d09-920e-b373d31f655d"
	}
	ctx := context.Background()

	// Check if path exists
	var pathCount int
	err = db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_path WHERE path_id = @p1`, pathID).Scan(&pathCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error on learning_path: %v\n", err)
		os.Exit(1)
	}

	// Check node count
	var nodeCount int
	err = db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM node WHERE path_id = @p1`, pathID).Scan(&nodeCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error on node: %v\n", err)
		os.Exit(1)
	}

	// Check path_enroll count
	var enrollCount int
	err = db.GetDB().QueryRowContext(ctx, `SELECT COUNT(*) FROM path_enroll WHERE path_id = @p1`, pathID).Scan(&enrollCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error on path_enroll: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== DATA PERSISTENCE CHECK ===\n")
	fmt.Printf("Path ID checked: %s\n", pathID)
	fmt.Printf("Paths found:     %d\n", pathCount)
	fmt.Printf("Nodes found:     %d\n", nodeCount)
	fmt.Printf("Enrollments found: %d\n", enrollCount)
	fmt.Printf("\nResult: ")
	if pathCount > 0 && nodeCount == 5 {
		fmt.Printf("✓ TEST DATA PERSISTED SUCCESSFULLY\n")
	} else {
		fmt.Printf("✗ TEST DATA DID NOT PERSIST\n")
		fmt.Printf("\nThis suggests the data was created in a transaction that was rolled back or isolated.\n")
	}
}
