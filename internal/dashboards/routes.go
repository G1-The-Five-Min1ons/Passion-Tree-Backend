package dashboard

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"passiontree/internal/connection"
	"passiontree/internal/dashboards/handler"
	"passiontree/internal/dashboards/repository"
	"passiontree/internal/dashboards/service"
	"passiontree/internal/pkg/jwt"
	"passiontree/internal/pkg/middleware"
)

func RegisterRoutes(r fiber.Router, db connection.Database, jwtService *jwt.Service, logger *slog.Logger) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, logger)
	h := handler.NewHandler(svc, logger)

	protected := r.Group("/dashboard", middleware.JWTMiddleware(jwtService, logger))
	{
		protected.Get("", h.GetDashboard)
	}
}
