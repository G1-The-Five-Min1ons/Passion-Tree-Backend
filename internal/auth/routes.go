package auth

import (
	"passiontree/internal/auth/handler"
	"passiontree/internal/auth/repository"
	"passiontree/internal/auth/service"
	"passiontree/internal/database"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, db database.Database) {
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	h := handler.NewHandler(svc)

	auth := r.Group("/auth")
	{
		// Public routes - no authentication required
		auth.Post("/register", h.Register)
		auth.Post("/login", h.Login)

		// Protected routes - require JWT authentication
		auth.Get("/profile", middleware.JWTMiddleware(), h.GetUserProfile)
		auth.Put("/profile", middleware.JWTMiddleware(), h.UpdateProfile)
		auth.Put("/user", middleware.JWTMiddleware(), h.UpdateUser)
		auth.Delete("/user", middleware.JWTMiddleware(), h.DeleteUser)
	}
}
