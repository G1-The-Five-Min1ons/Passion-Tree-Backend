package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetAllUsers returns all users (admin only)
func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.userSvc.GetAllUsers(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get all users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve users",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Users retrieved successfully",
		"data":    users,
	})
}

// GetDashboardStats returns dashboard statistics (admin only)
func (h *Handler) GetDashboardStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.userSvc.GetDashboardStats(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get dashboard stats", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve dashboard stats",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Dashboard stats retrieved successfully",
		"data":    stats,
	})
}
