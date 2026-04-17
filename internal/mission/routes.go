package mission

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	repoUser "passiontree/internal/auth/repository"
	"passiontree/internal/connection"
	"passiontree/internal/mission/handler"
	"passiontree/internal/mission/repository"
	"passiontree/internal/mission/service"
	"passiontree/internal/pkg/jwt"
	"passiontree/internal/pkg/middleware"
)

func RegisterRoutes(r fiber.Router, db connection.Database, jwtService *jwt.Service, logger *slog.Logger) {
	repo := repository.NewRepository(db)
	repouser := repoUser.NewRepository(db)
	svc := service.NewService(repo, repouser, logger)
	h := handler.NewHandler(svc, logger)

	protected := r.Group("/", middleware.JWTMiddleware(jwtService, logger))

	adminGroup := protected.Group("/admin/missions")
	{
		adminGroup.Post("/templates", h.CreateTemplate)
	}

	userGroup := protected.Group("/user/missions")
	{
		userGroup.Get("", h.GetMyMissions)
	}

	systemGroup := protected.Group("/system/missions")
	{
		systemGroup.Post("/auto-assign", h.TriggerAutoAssign)
	}
}
