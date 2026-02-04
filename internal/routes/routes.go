package routes

import (
	"log/slog"

	"passiontree/internal/database"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"

	auth "passiontree/internal/auth"
	history "passiontree/internal/history"
	learningpath "passiontree/internal/learning-path"
	reflection "passiontree/internal/reflection"
	resume "passiontree/internal/resume"

	"github.com/gofiber/fiber/v2"
)

// Setup configures all routes for the application
func Setup(app *fiber.App, db database.Database, aiClient *aiclient.AIClient, storageClient *storage.BlobService, logger *slog.Logger) {
	// Health check endpoint
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return healthCheck(c, db, storageClient)
	})

	auth.RegisterRoutes(api, db)
	learningpath.RegisterRoutes(api, db, aiClient, logger, storageClient)
	reflection.RegisterRoutes(api, db, aiClient, logger)
	history.RegisterRoutes(api, db)
	resume.RegisterRoutes(api, db)
}

// healthCheck returns the service health status
func healthCheck(c *fiber.Ctx, db database.Database, storageClient *storage.BlobService) error {
	response := fiber.Map{
		"status":  "up",
		"service": "Go Backend Orchestrator",
	}

	// Check database connection
	if err := db.CheckConnection(); err != nil {
		response["database"] = "disconnected"
		response["database_error"] = err.Error()
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	response["database"] = "connected"

	// Check Azure Storage connection
	if storageClient != nil {
		response["azure_storage"] = "connected"
	} else {
		response["azure_storage"] = "not_configured"
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
