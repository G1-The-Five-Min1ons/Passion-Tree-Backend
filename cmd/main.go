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

	app := fiber.New(fiber.Config{
		AppName: "Passion Tree Backend v1.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Setup routes with database instance
	routes.Setup(app, db)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	log.Printf("Starting Fiber server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
