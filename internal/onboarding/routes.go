package onboarding

import (
	"log/slog"

	"passiontree/internal/connection"
	"passiontree/internal/onboarding/handler"
	"passiontree/internal/onboarding/repository"
	"passiontree/internal/onboarding/service"
	"passiontree/internal/pkg/jwt"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, db connection.Database, recomputer service.RecommendationRecomputer, jwtService *jwt.Service, logger *slog.Logger) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, logger)
	if recomputer != nil {
		if setter, ok := svc.(service.RecomputerSetter); ok {
			setter.SetRecommendationRecomputer(recomputer)
		}
	}
	h := handler.NewHandler(svc, logger)

	protected := r.Group("/onboarding", middleware.JWTMiddleware(jwtService, logger))
	{
		protected.Post("", h.SaveOnboarding)
		protected.Get("", h.GetOnboarding)
	}
}
