package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetStats returns reflection statistics (admin only)
func (h *Handler) GetStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	stats, err := h.reflectSvc.GetReflectionStats(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection stats retrieved successfully")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection stats retrieved successfully",
		"data":    stats,
	})
}
