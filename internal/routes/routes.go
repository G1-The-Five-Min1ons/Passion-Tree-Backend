package routes

import (
	"log/slog"

	"passiontree/internal/connection"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/worker"

	"passiontree/internal/config"
	"passiontree/internal/pkg/jwt"

	auth "passiontree/internal/auth"
	learningpath "passiontree/internal/learning-path"
	onboarding "passiontree/internal/onboarding"
	recommendation "passiontree/internal/recommendation"
	reflection "passiontree/internal/reflection"
	setting "passiontree/internal/setting"
	upload "passiontree/internal/upload"
	dashboard "passiontree/internal/dashboards"

	"github.com/gofiber/fiber/v2"
)

// Setup configures all routes for the application
func Setup(app *fiber.App, db connection.Database, aiClient *aiclient.AIClient, storageClient *storage.BlobService, notificationWorker *worker.EmailNotificationWorker, logger *slog.Logger) {
	// Health check endpoint
	api := app.Group("/api/v1")

	// Load configuration for JWT service
	cfg, err := config.LoadDBConfig()
	if err != nil {
		logger.Error("startup_failed", "error", err)
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize JWT service
	jwtService := jwt.NewService(cfg)

	api.Get("/health", func(c *fiber.Ctx) error {
		return healthCheck(c, db, storageClient)
	})

	auth.RegisterRoutes(api, db, logger)
	learningpath.RegisterRoutes(api, db, aiClient, jwtService, logger, storageClient)
	reflection.RegisterRoutes(api, db, aiClient, jwtService, logger)
	upload.RegisterRoutes(api, jwtService, logger, storageClient)
	recommendation.RegisterRoutes(api, db, aiClient, jwtService, logger, storageClient)
	setting.RegisterRoutes(api, db, jwtService, notificationWorker, logger)
	dashboard.RegisterRoutes(api, db, jwtService, logger)
	onboarding.RegisterRoutes(api, db, jwtService, logger)
}

// healthCheck returns the service health status
func healthCheck(c *fiber.Ctx, db connection.Database, storageClient *storage.BlobService) error {
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
