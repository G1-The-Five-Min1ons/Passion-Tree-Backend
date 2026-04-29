package handler

import (
	"context"
	"strings"
	"time"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

const maintenanceTaskTimeout = 10 * time.Minute

// Search godoc
// @Summary      Semantic search learning paths
// @Description  Sends the query to the AI service which performs vector search over the learning-path collection.
// @Tags         Search
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.SearchPathRequest  true  "Search query"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Failure      502   {object}  apidoc.ErrorResponse  "AI service error"
// @Router       /learningpaths/search [post]
func (h *Handler) Search(c *fiber.Ctx) error {
	var req model.SearchPathRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	req.Query = strings.TrimSpace(req.Query)
	h.logger.InfoContext(ctx, "received learning path search request", "query", req.Query)

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

// DebugCollection godoc
// @Summary      Get vector collection debug info
// @Description  Returns size, vector count, and config for the named Qdrant collection. Defaults to learning_paths.
// @Tags         Search
// @Produce      json
// @Security     BearerAuth
// @Param        collection_name  path      string  true  "Qdrant collection name"
// @Success      200              {object}  apidoc.SuccessResponse
// @Failure      401              {object}  apidoc.ErrorResponse
// @Failure      502              {object}  apidoc.ErrorResponse
// @Router       /learningpaths/debug/collection/{collection_name} [get]
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

// BulkSyncLearningPaths godoc
// @Summary      Bulk sync all learning paths to Qdrant (admin)
// @Description  Starts an async job that pushes every learning path in SQL to the vector DB. Returns a task_id to poll.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      202  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse
// @Failure      409  {object}  apidoc.ErrorResponse  "Another bulk sync is already running"
// @Router       /learningpaths/sync/bulk [post]
func (h *Handler) BulkSyncLearningPaths(c *fiber.Ctx) error {
	const taskType = "bulk_sync"

	task, ok := h.tasks.create(taskType)
	if !ok {
		return h.handleError(c, apperror.NewConflict("A bulk sync task is already in progress"))
	}

	h.logger.InfoContext(c.UserContext(), "bulk sync started", "task_id", task.TaskID)

	go func(taskID string) {
		ctx, cancel := context.WithTimeout(context.Background(), maintenanceTaskTimeout)
		defer cancel()

		resp, err := h.searchSvc.BulkSyncLearningPaths(ctx)
		if err != nil {
			h.tasks.fail(taskID, err.Error())
			return
		}

		h.tasks.complete(taskID, resp)
	}(task.TaskID)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "Bulk sync started",
		"data": fiber.Map{
			"task_id": task.TaskID,
			"status":  task.Status,
		},
	})
}

// ReconcileLearningPaths godoc
// @Summary      Reconcile Qdrant with SQL (admin)
// @Description  Starts an async job that deletes stale vectors and syncs missing ones. Returns a task_id to poll.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      202  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse
// @Failure      409  {object}  apidoc.ErrorResponse
// @Router       /learningpaths/sync/reconcile [post]
func (h *Handler) ReconcileLearningPaths(c *fiber.Ctx) error {
	const taskType = "reconcile"

	task, ok := h.tasks.create(taskType)
	if !ok {
		return h.handleError(c, apperror.NewConflict("A reconcile task is already in progress"))
	}

	h.logger.InfoContext(c.UserContext(), "reconcile Qdrant↔SQL requested", "task_id", task.TaskID)

	go func(taskID string) {
		ctx, cancel := context.WithTimeout(context.Background(), maintenanceTaskTimeout)
		defer cancel()

		resp, err := h.searchSvc.ReconcileLearningPaths(ctx)
		if err != nil {
			h.tasks.fail(taskID, err.Error())
			h.logger.ErrorContext(ctx, "reconcile task failed", "task_id", taskID, "error", err)
			return
		}

		h.tasks.complete(taskID, resp)
		h.logger.InfoContext(ctx, "reconcile task completed", "task_id", taskID, "success", resp.Success)
	}(task.TaskID)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "Reconcile started",
		"data": fiber.Map{
			"task_id": task.TaskID,
			"status":  task.Status,
		},
	})
}

// GetSyncTaskStatus godoc
// @Summary      Poll an async sync/reconcile task (admin)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        task_id  path      string  true  "Task ID returned by /sync/bulk or /sync/reconcile"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      403      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/sync/tasks/{task_id} [get]
func (h *Handler) GetSyncTaskStatus(c *fiber.Ctx) error {
	taskID := strings.TrimSpace(c.Params("task_id"))
	if taskID == "" {
		return h.handleError(c, apperror.NewBadRequest("task_id is required"))
	}

	task, ok := h.tasks.get(taskID)
	if !ok {
		return h.handleError(c, apperror.NewNotFound("task not found"))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Task status retrieved successfully",
		"data":    task,
	})
}

// SyncLearningPath godoc
// @Summary      Sync a single learning path to Qdrant
// @Description  Pushes the specified learning path's data into the AI service's vector DB.
// @Tags         Search
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      502      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/sync/{path_id} [post]
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
