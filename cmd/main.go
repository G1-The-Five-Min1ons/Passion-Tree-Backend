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

const DefaultPort = "5000"

func main() {
	// Load configuration
	cfg, err := config.LoadDBConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// database connection with retry logic
	db, err := database.NewDatabaseWithRetry(cfg.DBConnString, 10, 3*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to database after multiple retries: %v", err)
	}
	defer db.Close()

	// Initialize AI client
	aiClient := aiclient.NewAIClient(cfg.AIServiceURL)
	log.Printf("AI Service URL: %s", cfg.AIServiceURL)

	// Initialize Azure Storage client (optional)
	var storageClient *database.StorageClient
	if cfg.AzureStorageConnString != "" {
		var err error
		storageClient, err = database.NewStorageClient(cfg)
		if err != nil {
			log.Printf("Warning: Failed to initialize Azure Storage client: %v", err)
		} else {
			log.Printf("Azure Storage client initialized successfully")
		}
	} else {
		log.Printf("Azure Storage not configured, skipping initialization")
	}

	app := fiber.New(fiber.Config{
		AppName: "Passion Tree Backend v1.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Setup routes with database instance, AI client, and storage client
	routes.Setup(app, db, aiClient, storageClient)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	log.Printf("Starting Fiber server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
