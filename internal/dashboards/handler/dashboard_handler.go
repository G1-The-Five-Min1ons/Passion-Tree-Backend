package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/middleware"
)

func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	dashboardData, err := h.dashboardSvc.GetDashboardData(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully retrieved dashboard data", "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Dashboard data retrieved successfully",
		"data":    dashboardData,
	})
}
