package handler

import (
	"context"
	"time"

	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// GetOnboarding godoc
// @Summary      Get the authenticated user's onboarding answers
// @Tags         Onboarding
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      404  {object}  apidoc.ErrorResponse
// @Router       /onboarding [get]
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
