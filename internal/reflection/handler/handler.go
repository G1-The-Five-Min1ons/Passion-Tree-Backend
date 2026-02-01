package handler

import (
	"log/slog"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/service"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	reflectSvc service.ReflectionService
	logger     *slog.Logger
}

func NewHandler(svc service.ReflectionService, logger *slog.Logger) *Handler {
	return &Handler{
		reflectSvc: svc,
		logger:     logger,
	}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	ctx := c.UserContext()

	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			h.logger.WarnContext(ctx, "application error",
				"code", appErr.Code,
				"message", appErr.Message,
				"cause", appErr.Log,
			)
		}
		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
		})
	}

	h.logger.ErrorContext(ctx, "unexpected system error", 
		"error", err.Error(),
		"path", c.Path(),
	)
	
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}
