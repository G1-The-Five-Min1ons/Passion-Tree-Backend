package auth

import (
	"passiontree/internal/auth/handler"
	"passiontree/internal/auth/repository"
	"passiontree/internal/auth/service"
	"passiontree/internal/database"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, db database.Database) {
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	h := handler.NewHandler(svc)

	auth := r.Group("/auth")
	{
		auth.Post("/register", h.Register)
		auth.Post("/login", h.Login)
		auth.Get("/profile", h.GetUserProfile)
		auth.Put("/profile", h.UpdateProfile)
		auth.Put("/user", h.UpdateUser)
		auth.Delete("/user", h.DeleteUser)
	}
}
