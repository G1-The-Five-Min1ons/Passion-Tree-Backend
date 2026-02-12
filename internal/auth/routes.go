package auth

import (
	"log/slog"
	
	"passiontree/internal/auth/handler"
	"passiontree/internal/auth/repository"
	"passiontree/internal/auth/service"
	"passiontree/internal/config"
	"passiontree/internal/connection"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, db connection.Database, logger *slog.Logger) {
	// Load configuration for email service
	cfg, err := config.LoadDBConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db.GetDB())

	// Initialize services with email configuration
	userSvc := service.NewUserServiceWithEmail(userRepo, tokenRepo, cfg, logger)

	h := handler.NewHandler(userSvc, logger)

	auth := r.Group("/auth")
	{
		// Public routes - no authentication required
		auth.Post("/register", h.Register)
		auth.Post("/login", middleware.RateLimitMiddleware(), h.Login)
		auth.Post("/verify-email", h.VerifyEmail)
		auth.Post("/resend-verification", h.ResendVerificationEmail)
		auth.Post("/forgot-password", h.ForgotPassword)
		auth.Post("/reset-password", h.ResetPassword)

		// Protected routes - require JWT authentication
		auth.Get("/profile", middleware.JWTMiddleware(), h.GetUserProfile)
		auth.Put("/profile", middleware.JWTMiddleware(), h.UpdateProfile)
		auth.Put("/user", middleware.JWTMiddleware(), h.UpdateUser)
		auth.Put("/change-password", middleware.JWTMiddleware(), h.ChangePassword)
		auth.Delete("/user", middleware.JWTMiddleware(), h.DeleteUser)
	}
}
