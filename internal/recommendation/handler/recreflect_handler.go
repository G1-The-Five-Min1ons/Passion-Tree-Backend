package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetRecommendations(c *fiber.Ctx) error {
	user_id := c.Query("user_id")
	path_id := c.Query("path_id")

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "requesting learning path recommendations", "user_id", user_id, "path_id", path_id)

	response, err := h.recreflectSvc.RecommendPathsForUser(ctx, path_id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Personalized recommendations generated successfully",
		"data":    response,
	})
}