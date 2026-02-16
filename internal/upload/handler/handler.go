package handler

import (
	"log/slog"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/upload/service"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	logger  *slog.Logger
	service service.Service
}

func NewHandler(logger *slog.Logger, service service.Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	ctx := c.UserContext()

	// Common Attributes
	logAttrs := []any{
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"user_agent", c.Get("User-Agent"),
		"request_id", c.GetRespHeader("X-Request-ID"), // Request ID Middleware
	}

	// Business/Handled Error)
	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			h.logger.WarnContext(ctx, "application handled error",
				append(logAttrs,
					"code", appErr.Code,
					"message", appErr.Message,
					"cause", appErr.Log,
				)...,
			)
		}

		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
		})
	}

	// System Error (Unhandled Error)
	h.logger.ErrorContext(ctx, "unhandled system error",
		append(logAttrs, "error", err.Error())...,
	)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}
