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
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "received learning path search request", "query", req.Query)

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if req.Query == "" {
		return h.handleError(c, apperror.NewBadRequest("search query is required"))
	}

	response, err := h.searchSvc.SearchLearningPaths(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "search completed successfully", "results_count", response.Total)

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

	h.logger.InfoContext(c.UserContext(), "retrieved collection info successfully", "collection", collectionName)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Collection info retrieved successfully",
		"data":    info,
	})
}

// BulkSyncLearningPaths pushes every path in SQL to Qdrant in one call.
func (h *Handler) BulkSyncLearningPaths(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Minute)
	defer cancel()

	h.logger.InfoContext(ctx, "bulk sync learning paths requested")

	resp, err := h.searchSvc.BulkSyncLearningPaths(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": resp.Success,
		"message": resp.Message,
		"data":    resp,
	})
}

// ReconcileLearningPaths aligns Qdrant with SQL (delete stale, sync missing).
func (h *Handler) ReconcileLearningPaths(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Minute)
	defer cancel()

	h.logger.InfoContext(ctx, "reconcile Qdrant↔SQL requested")

	resp, err := h.searchSvc.ReconcileLearningPaths(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": resp.Success,
		"message": resp.Message,
		"data":    resp,
	})
}

// SyncLearningPath syncs a single learning path from Azure DB to Qdrant
func (h *Handler) SyncLearningPath(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	if pathID == "" {
		return h.handleError(c, apperror.NewBadRequest("path_id is required"))
	}

	// Create context with timeout for sync operation
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	// Call search service to sync the learning path
	response, err := h.searchSvc.SyncLearningPath(ctx, pathID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "sync learning path to vector db operation successful", "path_id", pathID, "sync_status", response.Success)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": response.Success,
		"message": response.Message,
		"data": fiber.Map{
			"path_id": response.PathID,
		},
	})
}
