package handler

import (
	"context"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreateTree handles the creation of a new tree
func (h *Handler) CreateTree(c *fiber.Ctx) error {
	var req model.CreateTreeRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	resp, err := h.reflectSvc.CreateTree(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree created successfully", "tree_id", resp.TreeID, "album_id", req.AlbumID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "tree created successfully",
		"data":    resp,
	})
}

// GetTreeByID handles retrieving a tree by its ID
func (h *Handler) GetTreeByID(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "fetching tree details", "tree_id", treeID)

	tree, err := h.reflectSvc.GetTreeByID(ctx, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree retrieved successfully",
		"data": fiber.Map{
			"tree": tree,
		},
	})
}

// GetTreesByAlbumID handles retrieving all trees for an album
func (h *Handler) GetTreesByAlbumID(c *fiber.Ctx) error {
	albumID := c.Query("album_id")
	includeNodes := c.QueryBool("include_nodes", false)
	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	if albumID == "" {
		return h.handleError(c, apperror.NewBadRequest("album_id is required as query parameter"))
	}

	trees, err := h.reflectSvc.GetTreesByAlbumID(ctx, albumID, includeNodes)
	if err != nil {
		return h.handleError(c, err)
	}

	// Count handling based on type
	var count int
	switch v := trees.(type) {
	case []model.Tree:
		count = len(v)
	case []model.TreeResponse:
		count = len(v)
	}

	h.logger.InfoContext(ctx, "successfully retrieved trees for album", "album_id", albumID, "count", count, "include_nodes", includeNodes)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "trees retrieved successfully",
		"data": fiber.Map{
			"trees": trees,
			"count": count,
		},
	})
}

// UpdateTree handles updating an existing tree
func (h *Handler) UpdateTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateTreeRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	err := h.reflectSvc.UpdateTree(ctx, treeID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree updated successfully", "tree_id", treeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree updated successfully",
		"data": fiber.Map{
			"tree_id": treeID,
		},
	})
}

// DeleteTree handles deleting a tree
func (h *Handler) DeleteTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	err := h.reflectSvc.DeleteTree(ctx, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree deleted successfully", "tree_id", treeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree deleted successfully",
		"data": fiber.Map{
			"tree_id": treeID,
		},
	})
}

// PauseTree handles toggling pause/unpause state of a tree
func (h *Handler) PauseTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	var req model.PauseTreeRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	// Body is optional since we're toggling
	_ = c.BodyParser(&req)

	isPause, err := h.reflectSvc.PauseTree(ctx, treeID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	statusMsg := "paused"
	if !isPause {
		statusMsg = "unpaused"
	}

	h.logger.InfoContext(ctx, "tree "+statusMsg+" successfully", "tree_id", treeID, "is_pause", isPause)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree " + statusMsg + " successfully",
		"data": fiber.Map{
			"tree_id":  treeID,
			"is_pause": isPause,
		},
	})
}
