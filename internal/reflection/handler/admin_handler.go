package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetStats godoc
// @Summary      Get reflection statistics
// @Description  Returns aggregate reflection counts and trends across the platform.
// @Tags         Reflections
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /reflections/stats [get]
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
