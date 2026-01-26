package handler

import (
	"context"
	"time"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"

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

	trees, err := h.reflectSvc.GetTreesByAlbumID(ctx, albumID)
	if err != nil {
		return h.handleError(c, err)
	}

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

	err := h.reflectSvc.UpdateTree(ctx, treeID, req)
	if err != nil {
		return h.handleError(c, err)
	}

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

	err := h.reflectSvc.DeleteTree(ctx, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "tree deleted successfully",
	})
}
