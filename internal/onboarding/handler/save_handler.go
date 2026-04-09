package handler

import (
	"context"
	"time"

	"passiontree/internal/onboarding/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

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
