package auth

import (
	"log/slog"

	"passiontree/internal/auth/handler"
	"passiontree/internal/auth/repository"
	"passiontree/internal/auth/service"
	"passiontree/internal/config"
	"passiontree/internal/connection"
	"passiontree/internal/pkg/jwt"
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

	repo := repository.NewRepository(db)
	jwtService := jwt.NewService(cfg)
	emailSvc := service.NewEmailService(cfg, logger)
	userSvc := service.NewUserService(repo, emailSvc, cfg, jwtService, logger)
	socialAuthSvc := service.NewSocialAuthService(repo, cfg, jwtService, logger)
	h := handler.NewHandler(userSvc, socialAuthSvc, logger)

	auth := r.Group("/auth")
	{
		// Public routes - no authentication required
		auth.Post("/register", h.Register)
		auth.Post("/login", middleware.RateLimitMiddleware(), h.Login)
		auth.Post("/refresh", h.RefreshToken) // Refresh access token
		auth.Post("/verify-email", h.VerifyEmail)
		auth.Post("/resend-verification", h.ResendVerificationEmail)
		auth.Post("/forgot-password", h.ForgotPassword)
		auth.Post("/reset-password", h.ResetPassword)

		// Social Auth routes
		auth.Get("/google", h.GoogleLogin)
		auth.Get("/google/callback", h.GoogleCallback)
		auth.Get("/discord", h.DiscordLogin)
		auth.Get("/discord/callback", h.DiscordCallback)

		// Account Linking Confirmation
		auth.Post("/confirm-link", h.ConfirmAccountLink)

		// Native SSO route (for Android/mobile apps)
		auth.Post("/native/google", h.NativeGoogleSignIn)
	}

	// --- Protected Routes (Require JWT) ---
	protected := auth.Group("/", middleware.JWTMiddleware(jwtService, logger))
	{
		// Authentication
		protected.Post("/logout", h.Logout) // Logout and revoke tokens

		// Multi-device Session Management
		protected.Get("/sessions", h.GetActiveSessions)            // List all active sessions/devices
		protected.Delete("/sessions/:session_id", h.LogoutSession) // Logout from a specific device

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
