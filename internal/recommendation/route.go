package recommendation

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"

	"passiontree/internal/connection"
	"passiontree/internal/pkg/jwt"
	"passiontree/internal/pkg/middleware"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/handler"
	"passiontree/internal/recommendation/service"

	pathrepo "passiontree/internal/learning-path/repository"
	recrepo "passiontree/internal/recommendation/repository"
)

func RegisterRoutes(r fiber.Router, db connection.Database, aiClient *aiclient.AIClient, jwtService *jwt.Service, logger *slog.Logger, storageClient *storage.BlobService) {
	recRepository := recrepo.NewRepository(db)
	pathRepository := pathrepo.NewRepository(db)
	svc := service.NewService(recRepository, pathRepository, aiClient, logger)
	h := handler.NewHandler(svc, logger, storageClient)

	protected := r.Group("/", middleware.JWTMiddleware(jwtService, logger))

	paths := protected.Group("/reflect/recommendation")
	{
		paths.Get("", h.GetRecommendations)
	}
	homePaths := protected.Group("/home/recommendation")
	{
		homePaths.Get("", h.GetHomeRecommendations)
	}
	batchPaths := protected.Group("/batch/recommendation")
	{
		batchPaths.Post("/trigger", h.TriggerBatchRecommendation)
	}
}
