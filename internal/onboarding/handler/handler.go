package handler

import (
	"log/slog"

	"passiontree/internal/onboarding/service"
	"passiontree/internal/pkg/apperror"

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

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	ctx := c.UserContext()

	logAttrs := []any{
		"method",     c.Method(),
		"path",       c.Path(),
		"ip",         c.IP(),
		"user_agent", c.Get("User-Agent"),
		"request_id", c.GetRespHeader("X-Request-ID"),
	}

	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			h.logger.WarnContext(ctx, "application handled error",
				append(logAttrs,
					"code",    appErr.Code,
					"message", appErr.Message,
					"cause",   appErr.Log,
				)...,
			)
		}
		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
		})
	}

	h.logger.ErrorContext(ctx, "unhandled system error",
		append(logAttrs, "error", err.Error())...,
	)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}
