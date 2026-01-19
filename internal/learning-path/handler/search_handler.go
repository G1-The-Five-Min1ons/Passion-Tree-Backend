package handler

import (
	"context"
	"time"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

// Search handles search learning paths via AI service
func (h *Handler) Search(c *fiber.Ctx) error {
	var req model.SearchPathRequest
	// AI search ใช้เวลานาน ตั้ง timeout 30 วินาที
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// Validate query
	if req.Query == "" {
		return h.handleError(c, apperror.NewBadRequest("search query is required"))
	}

	// Call search service
	response, err := h.searchSvc.SearchLearningPaths(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Search completed successfully",
		"data":    response,
	})
}
