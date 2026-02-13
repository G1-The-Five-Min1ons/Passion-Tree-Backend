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
        logger.Error("startup_failed", "error", err) 
        panic("Failed to load configuration: " + err.Error())
    }

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db.GetDB())
	socialAuthRepo := repository.NewSocialAuthRepository(db.GetDB())

	// Initialize services with email configuration
	userSvc := service.NewUserServiceWithEmail(userRepo, tokenRepo, cfg, logger)
	socialAuthSvc := service.NewSocialAuthService(userRepo, socialAuthRepo, cfg, logger)

	// Initialize handlers (with social auth support)
	h := handler.NewHandlerWithSocialAuth(userSvc, socialAuthSvc, logger)

	auth := r.Group("/auth")
	{
		// Public routes - no authentication required
		auth.Post("/register", h.Register)
		auth.Post("/login", middleware.RateLimitMiddleware(), h.Login)
		auth.Post("/verify-email", h.VerifyEmail)
		auth.Post("/resend-verification", h.ResendVerificationEmail)
		auth.Post("/forgot-password", h.ForgotPassword)
		auth.Post("/reset-password", h.ResetPassword)

		// Social Auth routes
		auth.Get("/google", h.GoogleLogin)
		auth.Get("/google/callback", h.GoogleCallback)
		auth.Get("/discord", h.DiscordLogin)
		auth.Get("/discord/callback", h.DiscordCallback)

		// Native SSO route (for Android/mobile apps)
		auth.Post("/native/google", h.NativeGoogleSignIn)
	}

	// --- Protected Routes (Require JWT) ---
	protected := auth.Group("/", middleware.JWTMiddleware(logger))
	{
		// Profile & User Management
		protected.Get("/profile", h.GetUserProfile)
		protected.Put("/profile", h.UpdateProfile)
		protected.Put("/user", h.UpdateUser)
		protected.Put("/change-password", h.ChangePassword)
		protected.Delete("/user", h.DeleteUser)

		// Admin Routes (JWT + RBAC)
		adminOnly := protected.Group("/admin", middleware.RbacMiddleware(logger, "admin"))
		{
			adminOnly.Get("/dashboard", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"message": "Welcome to Admin Dashboard"})
			})
			// สามารถเพิ่ม Route สำหรับจัดการ User ในนี้ได้
		}

		// Teacher Routes
		teacherOnly := protected.Group("/teacher", middleware.RbacMiddleware(logger, "teacher"))
		{
			teacherOnly.Get("/dashboard", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"message": "Welcome Teacher"})
			})
		}
	}
}
