package handler

import (
	"log/slog"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/setting/service"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc    service.Service
	logger *slog.Logger
}

func NewHandler(svc service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

// handleError provides centralized error handling similar to Learning-Path
func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	ctx := c.UserContext()

	logAttrs := []any{
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"user_agent", c.Get("User-Agent"),
		"request_id", c.GetRespHeader("X-Request-ID"),
	}

	// Handle business/known errors
	if appErr, ok := err.(*apperror.AppError); ok {
		h.logger.WarnContext(ctx, appErr.Message, logAttrs...)
		return c.Status(appErr.Code).JSON(fiber.Map{
			"error":      appErr.Message,
			"error_code": appErr.Code,
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}

	// Handle unexpected errors
	h.logger.ErrorContext(ctx, "internal_server_error", append(logAttrs, "error", err.Error())...)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":      "Internal server error",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// successResponse returns a JSON success response
func (h *Handler) successResponse(message string) fiber.Map {
	return fiber.Map{"message": message}
}
