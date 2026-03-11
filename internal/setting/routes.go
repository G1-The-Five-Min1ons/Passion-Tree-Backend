package setting

import (
	"log/slog"
	"os"

	"passiontree/internal/connection"
	"passiontree/internal/pkg/jwt"
	"passiontree/internal/pkg/middleware"
	"passiontree/internal/setting/handler"
	"passiontree/internal/setting/repository"
	"passiontree/internal/setting/service"
	"passiontree/internal/worker"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, db connection.Database, jwtService *jwt.Service, notificationWorker *worker.EmailNotificationWorker, logger *slog.Logger) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, logger)
	h := handler.NewHandler(svc, logger)

	settings := r.Group("/settings", middleware.JWTMiddleware(jwtService, logger))
	{
		// Get user settings
		settings.Get("/", h.GetSettings)

		// Update user settings
		settings.Put("/", h.UpdateSettings)

		// Get specific setting
		settings.Get("/:key", h.GetSetting)

		// Update specific setting
		settings.Put("/:key", h.UpdateSetting)

		// Delete setting
		settings.Delete("/:key", h.DeleteSetting)
	}

	if os.Getenv("APP_ENV") != "production" && notificationWorker != nil {
		debug := r.Group("/debug")
		debug.Post("/notifications/daily", func(c *fiber.Ctx) error {
			go notificationWorker.RunDailyNotifications()
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success": true,
				"message": "daily notifications triggered",
			})
		})

		debug.Post("/notifications/weekly", func(c *fiber.Ctx) error {
			go notificationWorker.RunWeeklyNotifications()
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success": true,
				"message": "weekly notifications triggered",
			})
		})

		logger.Info("debug notification trigger endpoints enabled")
	}
}
