package handler

import (
	"context"
	"time"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetUserHistory(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
    defer cancel()

	historyList, err := h.historySvc.GetUserHistory(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "retrieved user history successfully", "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User history retrieved successfully",
		"data":    historyList,
	})
}