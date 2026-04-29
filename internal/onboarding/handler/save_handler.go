package handler

import (
	"context"
	"time"

	"passiontree/internal/onboarding/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// SaveOnboarding godoc
// @Summary      Save onboarding answers
// @Description  Persists the user's onboarding survey answers (interests, goals, etc.).
// @Tags         Onboarding
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.SaveOnboardingRequest  true  "Onboarding answers"
// @Success      200   {object}  apidoc.MessageResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /onboarding [post]
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
