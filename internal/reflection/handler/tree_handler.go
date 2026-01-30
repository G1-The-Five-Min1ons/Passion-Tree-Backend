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

	h.logger.InfoContext(ctx, "creating new reflection tree", "album_id", req.AlbumID, "title", req.Title)

	resp, err := h.reflectSvc.CreateTree(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree created successfully", "tree_id", resp.TreeID, "album_id", req.AlbumID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "tree created successfully",
		"data": fiber.Map{
			"tree": resp,
		},
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

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"tree": tree,
		},
	})
}

// GetTreesByAlbumID handles retrieving all trees for an album
func (h *Handler) GetTreesByAlbumID(c *fiber.Ctx) error {
	albumID := c.Query("album_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if albumID == "" {
		return h.handleError(c, apperror.NewBadRequest("album_id is required as query parameter"))
	}

	h.logger.InfoContext(ctx, "fetching all trees in album", "album_id", albumID)

	trees, err := h.reflectSvc.GetTreesByAlbumID(ctx, albumID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully retrieved trees for album", "album_id", albumID, "count", len(trees))

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"trees": trees,
			"count": len(trees),
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

	h.logger.InfoContext(ctx, "updating tree information", "tree_id", treeID)

	err := h.reflectSvc.UpdateTree(ctx, treeID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree updated successfully", "tree_id", treeID)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "tree updated successfully",
	})
}

// DeleteTree handles deleting a tree
func (h *Handler) DeleteTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "requesting tree deletion", "tree_id", treeID)

	err := h.reflectSvc.DeleteTree(ctx, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree deleted successfully", "tree_id", treeID)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "tree deleted successfully",
	})
}
