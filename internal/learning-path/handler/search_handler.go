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

// DebugCollection retrieves debug information about a collection from AI service
func (h *Handler) DebugCollection(c *fiber.Ctx) error {
	collectionName := c.Params("collection_name")
	if collectionName == "" {
		collectionName = "learning_paths"
	}

	// Call search service to get collection info
	info, err := h.searchSvc.GetCollectionInfo(collectionName)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Collection info retrieved successfully",
		"data":    info,
	})
}

// SyncLearningPath syncs a single learning path from Azure DB to Qdrant
func (h *Handler) SyncLearningPath(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	if pathID == "" {
		return h.handleError(c, apperror.NewBadRequest("path_id is required"))
	}

	// Call search service to sync the learning path
	response, err := h.searchSvc.SyncLearningPath(pathID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": response.Success,
		"message": response.Message,
		"data": fiber.Map{
			"path_id": response.PathID,
		},
	})
}
