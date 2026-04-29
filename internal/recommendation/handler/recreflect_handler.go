package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
)

// GetRecommendations godoc
// @Summary      Get learning-path recommendations from a reflection tree
// @Description  Returns personalized learning paths derived from the given reflection tree's content.
// @Tags         Recommendations
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  query     string  true  "Reflection tree ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /reflect/recommendation [get]
func (h *Handler) GetRecommendations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Please log in again to continue",
		})
	}
	treeID := c.Query("tree_id")

	if treeID == "" {
		return h.handleError(c, apperror.NewBadRequest("Missing required reflection data"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "received request for path recommendations", "user_id", userID, "tree_id", treeID)

	response, err := h.recSvc.RecommendPathsForUser(ctx, userID, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Your personalized recommendations are ready",
		"data":    response,
	})
}
