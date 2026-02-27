package upload

import (
	"log/slog"

	"passiontree/internal/pkg/jwt"
	// "passiontree/internal/pkg/middleware"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/upload/handler"
	"passiontree/internal/upload/service"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, jwtService *jwt.Service, logger *slog.Logger, storageClient *storage.BlobService) {
	svc := service.NewService(logger, storageClient)
	h := handler.NewHandler(logger, svc)

	// protected := r.Group("/", middleware.JWTMiddleware(jwtService, logger))
	protected := r.Group("/") //for dev

	group := protected.Group("/upload")
	{
		group.Post("/presignedimg-url", h.GetPresignedIMGURL)
	}
}
