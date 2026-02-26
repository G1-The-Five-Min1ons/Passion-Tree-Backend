package recommendation

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"

	"passiontree/internal/connection"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/handler"
	"passiontree/internal/recommendation/repository"
	"passiontree/internal/recommendation/service"
)

func RegisterRoutes(r fiber.Router, db connection.Database, aiClient *aiclient.AIClient, logger *slog.Logger, storageClient *storage.BlobService) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, aiClient, logger)
	h := handler.NewHandler(svc, logger, storageClient)

	paths := r.Group("/reflect/recomendation")
	{
		paths.Get("", h.GetRecommendations)
	}
}
