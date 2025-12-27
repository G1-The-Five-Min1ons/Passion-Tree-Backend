package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

//go:embed *.sql
var migrationFiles embed.FS

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB) error {
	// Create migrations table if not exists
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load migrations from embedded files
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if applied[migration.Version] {
			log.Printf("Migration %d already applied: %s", migration.Version, migration.Name)
			continue
		}

		log.Printf("Running migration %d: %s", migration.Version, migration.Name)
		if err := runMigration(db, migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
		}
		log.Printf("✅ Migration %d completed: %s", migration.Version, migration.Name)
	}

	log.Println("✅ All migrations completed successfully")
	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'schema_migrations')
	BEGIN
		CREATE TABLE schema_migrations (
			version INT PRIMARY KEY,
			name NVARCHAR(255) NOT NULL,
			applied_at DATETIME2 NOT NULL DEFAULT GETDATE()
		)
	END`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

func loadMigrations() ([]Migration, error) {
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrationFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		migration, err := parseMigration(entry.Name(), string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigration(filename, content string) (Migration, error) {
	// Extract version from filename (e.g., "001_create_users_table.sql" -> 1)
	var version int
	var name string
	_, err := fmt.Sscanf(filename, "%d_%s", &version, &name)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	// Split up and down migrations
	parts := strings.Split(content, "-- +migrate Down")
	if len(parts) != 2 {
		return Migration{}, fmt.Errorf("migration must have both Up and Down sections")
	}

	upSQL := strings.TrimPrefix(parts[0], "-- +migrate Up")
	downSQL := parts[1]

	return Migration{
		Version: version,
		Name:    strings.TrimSuffix(name, ".sql"),
		Up:      strings.TrimSpace(upSQL),
		Down:    strings.TrimSpace(downSQL),
	}, nil
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func runMigration(db *sql.DB, migration Migration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	_, err = tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, name) VALUES (@p1, @p2)",
		migration.Version, migration.Name)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}
