package handler

import (
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

// Search handles search learning paths via AI service
func (h *Handler) Search(c *fiber.Ctx) error {
	var req model.SearchPathRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// Validate query
	if req.Query == "" {
		return h.handleError(c, apperror.NewBadRequest("search query is required"))
	}

	// Call search service
	response, err := h.searchSvc.SearchLearningPaths(req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Search completed successfully",
		"data":    response,
	})
}
