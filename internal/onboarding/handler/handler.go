package handler

import (
	"context"
	"log/slog"
	"time"

	"passiontree/internal/onboarding/model"
	"passiontree/internal/onboarding/service"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

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

// POST /onboarding
func (h *Handler) SaveOnboarding(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.SaveOnboardingRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.svc.SaveOnboarding(ctx, userID, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "onboarding saved successfully", "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Onboarding saved successfully",
	})
}

// GET /onboarding
func (h *Handler) GetOnboarding(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	data, err := h.svc.GetOnboarding(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Onboarding data retrieved successfully",
		"data":    data,
	})
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
