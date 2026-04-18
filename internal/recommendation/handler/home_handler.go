package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/middleware"
)

func (h *Handler) GetHomeRecommendations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Please log in again to continue",
		})
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "received request for home path recommendations", "user_id", userID)

	response, err := h.recSvc.RecommendHomePathsForUser(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Recommendations retrieved successfully",
		"data":    response,
	})
}

func (h *Handler) TriggerBatchRecommendation(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Minute)
	defer cancel()

	h.logger.InfoContext(ctx, "manual trigger for recommendation batch started")

	err := h.recSvc.RunDailyRecommendationBatch(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Batch recommendation triggered and completed successfully!",
	})
}
