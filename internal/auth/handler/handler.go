package handler

import (
	"html/template"
	"log/slog"
	"os"

	"passiontree/internal/auth/service"
	"passiontree/internal/config"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	userSvc          service.UserService
	socialAuthSvc    service.SocialAuthService
	logger           *slog.Logger
	isProduction     bool
	config           *config.Config
	oauthRedirectTpl *template.Template
}

func NewHandler(userSvc service.UserService, socialAuthSvc service.SocialAuthService, cfg *config.Config, logger *slog.Logger) *Handler {
	appEnv := os.Getenv("APP_ENV")

	// Load OAuth redirect template
	tpl, err := template.ParseFiles("internal/auth/handler/templates/oauth_redirect.html")
	if err != nil {
		logger.Warn("failed to load oauth_redirect template, will use inline HTML as fallback", "error", err)
	}

	return &Handler{
		userSvc:          userSvc,
		socialAuthSvc:    socialAuthSvc,
		logger:           logger,
		isProduction:     appEnv == "production",
		config:           cfg,
		oauthRedirectTpl: tpl,
	}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		// use tagged switch seperate Log
		switch appErr.Code {
		case fiber.StatusInternalServerError:
			h.logger.ErrorContext(c.UserContext(), "server_error",
				"error", appErr.Log,
				"path", c.Path(),
			)

		case fiber.StatusTooManyRequests:
			h.logger.WarnContext(c.UserContext(), "auth_lockout",
				"error", appErr.Message,
				"ip", c.IP(),
			)

		case fiber.StatusUnauthorized, fiber.StatusForbidden:
			h.logger.WarnContext(c.UserContext(), "security_anomaly",
				"code", appErr.Code,
				"error", appErr.Message,
				"path", c.Path(),
				"ip", c.IP(),
			)

		// for 400, 401, 404
		default:
			// No-op: Just return the response
		}

		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
		})
	}

	// Unknown Error
	h.logger.ErrorContext(c.UserContext(), "unknown_error",
		"err", err,
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
	)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}

// isSecureCookie returns true if running in production environment
func (h *Handler) isSecureCookie() bool {
	return h.isProduction
}
