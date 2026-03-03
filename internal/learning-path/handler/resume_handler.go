package handler

import (
	"context"
	"time"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetResume(c *fiber.Ctx) error {
	pathID := c.Query("path_id")
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

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

	h.logger.InfoContext(ctx, "retrieved resume node successfully", "user_id", userID, "path_id", pathID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Resume node retrieved successfully",
		"data":    resp,
	})
}
