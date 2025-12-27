package routes

import (
	"passiontree/internal/database"

	"github.com/gofiber/fiber/v2"
)

// Setup configures all routes for the application
func Setup(app *fiber.App, db database.Database) {
	// Health check endpoint
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return healthCheck(c, db)
	})
}

// healthCheck returns the service health status
func healthCheck(c *fiber.Ctx, db database.Database) error {
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
	return c.Status(fiber.StatusOK).JSON(response)
}
