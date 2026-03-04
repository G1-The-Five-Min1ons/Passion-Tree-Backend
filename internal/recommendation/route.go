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
	"passiontree/internal/recommendation/repository"
	"passiontree/internal/recommendation/service"
)

func RegisterRoutes(r fiber.Router, db connection.Database, aiClient *aiclient.AIClient, jwtService *jwt.Service, logger *slog.Logger, storageClient *storage.BlobService) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, aiClient, logger)
	h := handler.NewHandler(svc, logger, storageClient)

	protected := r.Group("/", middleware.JWTMiddleware(jwtService, logger))

	paths := protected.Group("/reflect/recomendation")
	{
		paths.Get("", h.GetRecommendations)
	}
}
