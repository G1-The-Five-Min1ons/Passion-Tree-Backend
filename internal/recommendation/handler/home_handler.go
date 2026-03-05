package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/middleware"
)

func (h *Handler) GetHomeRecommendations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "received request for home path recommendations", "user_id", userID)

	response, err := h.recreflectSvc.RecommendHomePathsForUser(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Home recommendations generated successfully",
		"data":    response,
	})
}
