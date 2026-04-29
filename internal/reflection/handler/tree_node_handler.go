package handler

import (
	"context"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreateTreeNode godoc
// @Summary      Create a tree node
// @Tags         Tree Nodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTreeNodeRequest  true  "Tree node payload"
// @Success      201   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /tree-nodes [post]
func (h *Handler) CreateTreeNode(c *fiber.Ctx) error {
	var req model.CreateTreeNodeRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	resp, err := h.reflectSvc.CreateTreeNode(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree node created successfully", "tree_node_id", resp.TreeNodeID, "tree_id", req.TreeID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "tree node created successfully",
		"data": fiber.Map{
			"tree_node": resp,
		},
	})
}

// GetTreeNodesByTreeID godoc
// @Summary      List tree nodes belonging to a tree
// @Tags         Tree Nodes
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  query     string  true  "Tree ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /tree-nodes [get]
func (h *Handler) GetTreeNodesByTreeID(c *fiber.Ctx) error {
	treeID := c.Query("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if treeID == "" {
		return h.handleError(c, apperror.NewBadRequest("tree_id query parameter is required"))
	}

	h.logger.InfoContext(ctx, "fetching tree nodes", "tree_id", treeID)

	nodes, err := h.reflectSvc.GetTreeNodesByTreeID(ctx, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree nodes retrieved successfully",
		"data": fiber.Map{
			"tree_nodes": nodes,
			"count":      len(nodes),
		},
	})
}

// GetTreeNodeByID godoc
// @Summary      Get a tree node by ID
// @Tags         Tree Nodes
// @Produce      json
// @Security     BearerAuth
// @Param        tree_node_id  path      string  true  "Tree node ID"
// @Success      200           {object}  apidoc.SuccessResponse
// @Failure      401           {object}  apidoc.ErrorResponse
// @Failure      404           {object}  apidoc.ErrorResponse
// @Router       /tree-nodes/{tree_node_id} [get]
func (h *Handler) GetTreeNodeByID(c *fiber.Ctx) error {
	treeNodeID := c.Params("tree_node_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "fetching tree node details", "tree_node_id", treeNodeID)

	node, err := h.reflectSvc.GetTreeNodeByID(ctx, treeNodeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree node retrieved successfully",
		"data": fiber.Map{
			"tree_node": node,
		},
	})
}

// UpdateTreeNode godoc
// @Summary      Update a tree node
// @Tags         Tree Nodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        tree_node_id  path      string                       true  "Tree node ID"
// @Param        body          body      model.UpdateTreeNodeRequest  true  "Updated fields"
// @Success      200           {object}  apidoc.SuccessResponse
// @Failure      400           {object}  apidoc.ErrorResponse
// @Failure      401           {object}  apidoc.ErrorResponse
// @Failure      404           {object}  apidoc.ErrorResponse
// @Router       /tree-nodes/{tree_node_id} [put]
func (h *Handler) UpdateTreeNode(c *fiber.Ctx) error {
	treeNodeID := c.Params("tree_node_id")
	var req model.UpdateTreeNodeRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	err := h.reflectSvc.UpdateTreeNode(ctx, treeNodeID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree node updated successfully", "tree_node_id", treeNodeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree node updated successfully",
		"data": fiber.Map{
			"tree_node_id": treeNodeID,
		},
	})
}

// DeleteTreeNode godoc
// @Summary      Delete a tree node
// @Tags         Tree Nodes
// @Produce      json
// @Security     BearerAuth
// @Param        tree_node_id  path      string  true  "Tree node ID"
// @Success      200           {object}  apidoc.SuccessResponse
// @Failure      401           {object}  apidoc.ErrorResponse
// @Failure      404           {object}  apidoc.ErrorResponse
// @Router       /tree-nodes/{tree_node_id} [delete]
func (h *Handler) DeleteTreeNode(c *fiber.Ctx) error {
	treeNodeID := c.Params("tree_node_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	err := h.reflectSvc.DeleteTreeNode(ctx, treeNodeID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree node deleted successfully", "tree_node_id", treeNodeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree node deleted successfully",
		"data": fiber.Map{
			"tree_node_id": treeNodeID,
		},
	})
}
