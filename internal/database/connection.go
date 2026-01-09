package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// Connection pool configuration
const (
	MaxOpenConns      = 10
	MaxIdleConns      = 5
	ConnMaxLifetime   = 5 * time.Minute
	PingTimeout       = 5 * time.Second
	MaxRetryDelay     = 30 * time.Second
	BackoffMultiplier = 2
)

// Database interface for easy mocking in tests
type Database interface {
	GetDB() *sql.DB
	CheckConnection() error
	Close() error
}

// sqlDatabase implements the Database interface
type sqlDatabase struct {
	db *sql.DB
}

// NewDatabase creates a new database connection with the provided connection string
func NewDatabase(connString string) (Database, error) {
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Configure connection pool for optimal performance
	configureConnectionPool(db)

	// Verify connection
	if err := pingDatabase(db); err != nil {
		return nil, err
	}

	return &sqlDatabase{db: db}, nil
}

// NewDatabaseWithRetry creates a database connection with retry logic for cloud databases
// maxRetries: number of retry attempts (recommended: 5-10)
// initialDelay: initial delay before first retry (recommended: 2-5 seconds)
func NewDatabaseWithRetry(connString string, maxRetries int, initialDelay time.Duration) (Database, error) {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Database connection attempt %d/%d", attempt, maxRetries)

		db, err := NewDatabase(connString)
		if err == nil {
			log.Printf("Database connected successfully on attempt %d", attempt)
			return db, nil
		}

		lastErr = err
		log.Printf("Connection attempt %d failed: %v", attempt, err)

		if attempt < maxRetries {
			log.Printf("Retrying in %v...", delay)
			time.Sleep(delay)
			delay = calculateNextDelay(delay)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, lastErr)
}

// configureConnectionPool sets optimal connection pool settings
func configureConnectionPool(db *sql.DB) {
	db.SetMaxOpenConns(MaxOpenConns)
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetConnMaxLifetime(ConnMaxLifetime)
}

// pingDatabase verifies database connectivity
func pingDatabase(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), PingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// calculateNextDelay implements exponential backoff with max delay cap
func calculateNextDelay(currentDelay time.Duration) time.Duration {
	nextDelay := currentDelay * BackoffMultiplier
	if nextDelay > MaxRetryDelay {
		return MaxRetryDelay
	}
	return nextDelay
}

// GetDB returns the underlying sql.DB instance
func (s *sqlDatabase) GetDB() *sql.DB {
	return s.db
}

// CheckConnection verifies the database connection is still alive
func (s *sqlDatabase) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.db.PingContext(ctx)
}

// Close closes the database connection
func (s *sqlDatabase) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
