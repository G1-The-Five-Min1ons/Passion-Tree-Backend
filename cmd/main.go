package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/robfig/cron/v3"

	"passiontree/internal/config"
	"passiontree/internal/connection"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/pkg/logger"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/routes"
	"passiontree/internal/worker"
)

const (
	DefaultPort     = "5000"
	DBRetryAttempts = 10
	DBRetryDelay    = 3 * time.Second
	AppName         = "Passion Tree Backend v1.0"
)

func main() {
	isDev := os.Getenv("APP_ENV") != "production" 
	myLogger := logger.SetupLogger(isDev)

	// Load configuration
	cfg, err := config.LoadDBConfig()
    if err != nil {
        myLogger.Error("Failed to load config", "error", err)
        os.Exit(1)
    }

	// Initialize database connection
	db, err := initializeDatabase(cfg.DBConnString, myLogger)
    if err != nil {
        myLogger.Error("Failed to initialize database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

	// Initialize AI client
	aiClient := initializeAIClient(cfg.AIServiceURL, myLogger)

	// Initialize Azure Storage client
	storageClient := initializeStorageClient(cfg, myLogger)

	// Setup Fiber with custom Logger
	app := createFiberApp(myLogger)
    routes.Setup(app, db, aiClient, storageClient, myLogger)

	cronJob := initializeBackgroundJobs(db, storageClient, myLogger)
    defer cronJob.Stop()

	// Start server
	port := getPort()
	myLogger.Info("starting server", "port", port, "app_name", AppName)

	if err := app.Listen(":" + port); err != nil {
        myLogger.Error("server crashed", "error", err)
        os.Exit(1)
    }
}

// initializeDatabase connects to the database with retry logic
func initializeDatabase(connString string, logger *slog.Logger) (connection.Database, error) {
	db, err := connection.NewDatabaseWithRetry(connString, DBRetryAttempts, DBRetryDelay)
	if err != nil {
		return nil, err
	}
	logger.Info("database connected successfully")
	return db, nil
}

// initializeAIClient creates and configures the AI service client
func initializeAIClient(serviceURL string, logger *slog.Logger) *aiclient.AIClient {
	if serviceURL == "" {
		logger.Error("AI_SERVICE_URL is not set or empty. Please check your environment variables or configuration.")
		os.Exit(1)
	}
	client := aiclient.NewAIClient(serviceURL)
	logger.Info("AI Service configured", "service_url", serviceURL)
	return client
}

// initializeStorageClient creates Azure Storage client if configured
func initializeStorageClient(cfg *config.Config, logger *slog.Logger) *storage.BlobService {
	if cfg.AzureStorageConnString == "" {
		logger.Info("Azure Storage not configured, skipping initialization")
		return nil
	}

	storageClient, err := connection.InitBlobStorage(cfg)
	if err != nil {
		logger.Warn("Failed to initialize Azure Storage", "error", err)
		return nil
	}

	logger.Info("Azure Storage initialized successfully")
	return storageClient
}

// createFiberApp creates and configures the Fiber application with middleware
func createFiberApp(logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: AppName,
	})

	// Apply middleware
	app.Use(flogger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	return app
}

// getPort returns the server port from environment or default
func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return DefaultPort
}

func initializeBackgroundJobs(db connection.Database, storage *storage.BlobService, logger *slog.Logger) *cron.Cron {
    cleanupWorker := worker.NewCleanupWorker(db, storage)
    c := cron.New()
    
    // Run every midnight
    _, err := c.AddFunc("0 0 * * *", func() {
        cleanupWorker.RunCleanup()
    })

    if err != nil {
        logger.Warn("Error initializing background jobs: %v", err)
        return c
    }

    c.Start()
    logger.Warn("Background jobs started")
    return c
}
