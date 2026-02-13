package handler

import (
	"context"
	"time"

	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetResume(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	pathID := c.Query("path_id")

	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}
	if pathID == "" {
		return h.handleError(c, apperror.NewBadRequest("path_id is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
    defer cancel()
	
	resp, err := h.resumeSvc.GetResumeNode(ctx, userID, pathID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Resume node retrieved successfully",
		"data":    resp,
	})
}