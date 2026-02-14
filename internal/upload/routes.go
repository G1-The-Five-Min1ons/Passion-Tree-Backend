package upload

import (
	"log/slog"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/upload/handler"
	"passiontree/internal/upload/service"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, logger *slog.Logger, storageClient *storage.BlobService) {
	svc := service.NewService(logger, storageClient)
	h := handler.NewHandler(logger, svc)

	group := r.Group("/upload")
	{
		group.Post("/presignedimg-url", h.GetPresignedIMGURL)
	}
}
