package handler

import (
	"log/slog"

	"passiontree/internal/auth/service"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	userSvc service.UserService
	logger  *slog.Logger
}

func NewHandler(userSvc service.UserService, logger *slog.Logger) *Handler {
	return &Handler{
		userSvc: userSvc,
		logger:  logger,
	}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		// use tagged switch seperate Log 
		switch appErr.Code {
		case fiber.StatusInternalServerError:
			h.logger.ErrorContext(c.UserContext(), "internal server error",
				"msg", appErr.Log, 
				"path", c.Path(),
			)

		case fiber.StatusTooManyRequests:
			h.logger.WarnContext(c.UserContext(), "account rate limited",
				"msg", appErr.Message,
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

	h.logger.ErrorContext(c.UserContext(), "UNKNOWN ERR", "err", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}