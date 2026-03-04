package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
)

func (h *Handler) GetRecommendations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}
	treeID := c.Query("tree_id")

	if userID == "" || treeID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id and tree_id are required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "received request for path recommendations", "user_id", userID, "tree_id", treeID)

	response, err := h.recreflectSvc.RecommendPathsForUser(ctx, userID, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Personalized recommendations generated successfully",
		"data":    response,
	})
}
