package auth

import (
	"passiontree/internal/auth/handler"
	"passiontree/internal/auth/repository"
	"passiontree/internal/auth/service"
	"passiontree/internal/config"
	"passiontree/internal/database"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(r fiber.Router, db database.Database) {
	// Load configuration for email service
	cfg, err := config.LoadDBConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db.GetDB())

	// Initialize services
	userSvc := service.NewUserService(userRepo, tokenRepo)

	// Initialize email service if SMTP is configured
	if cfg.SMTPHost != "" {
		emailSvc := service.NewEmailService(cfg)
		// Set email service to user service
		if svc, ok := userSvc.(interface{ SetEmailService(service.EmailService) }); ok {
			svc.SetEmailService(emailSvc)
		}
	}

	h := handler.NewHandler(userSvc)

	auth := r.Group("/auth")
	{
		// Public routes - no authentication required
		auth.Post("/register", h.Register)
		auth.Post("/login", h.Login)
		auth.Get("/verify-email", h.VerifyEmail)
		auth.Post("/resend-verification", h.ResendVerificationEmail)

		// Protected routes - require JWT authentication
		auth.Get("/profile", middleware.JWTMiddleware(), h.GetUserProfile)
		auth.Put("/profile", middleware.JWTMiddleware(), h.UpdateProfile)
		auth.Put("/user", middleware.JWTMiddleware(), h.UpdateUser)
		auth.Delete("/user", middleware.JWTMiddleware(), h.DeleteUser)
	}
}
