package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"passiontree/internal/config"
	"passiontree/internal/database"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/routes"
)

const (
	DefaultPort     = "5000"
	DBRetryAttempts = 10
	DBRetryDelay    = 3 * time.Second
	AppName         = "Passion Tree Backend v1.0"
)

func main() {
	// Load configuration
	cfg, err := config.LoadDBConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := initializeDatabase(cfg.DBConnString)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize AI client
	aiClient := initializeAIClient(cfg.AIServiceURL)

	// Initialize Azure Storage client (optional)
	storageClient := initializeStorageClient(cfg)

	app := createFiberApp()
	routes.Setup(app, db, aiClient, storageClient)

	// Start server
	port := getPort()
	log.Printf("Starting server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

// initializeDatabase connects to the database with retry logic
func initializeDatabase(connString string) (database.Database, error) {
	db, err := database.NewDatabaseWithRetry(connString, DBRetryAttempts, DBRetryDelay)
	if err != nil {
		return nil, err
	}
	log.Println("Database connected successfully")
	return db, nil
}

// initializeAIClient creates and configures the AI service client
func initializeAIClient(serviceURL string) *aiclient.AIClient {
	client := aiclient.NewAIClient(serviceURL)
	log.Printf("AI Service configured: %s", serviceURL)
	return client
}

// initializeStorageClient creates Azure Storage client if configured
func initializeStorageClient(cfg *config.Config) *database.StorageClient {
	if cfg.AzureStorageConnString == "" {
		log.Println("Azure Storage not configured, skipping initialization")
		return nil
	}

	storageClient, err := database.NewStorageClient(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize Azure Storage: %v", err)
		return nil
	}

	log.Println("Azure Storage initialized successfully")
	return storageClient
}

// createFiberApp creates and configures the Fiber application with middleware
func createFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: AppName,
	})

	// Apply middleware
	app.Use(logger.New())
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