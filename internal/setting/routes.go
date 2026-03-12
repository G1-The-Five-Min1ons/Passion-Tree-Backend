package setting

import (
	"log/slog"

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
	registerDebugNotificationRoutes(r, db, notificationWorker, logger)
}
