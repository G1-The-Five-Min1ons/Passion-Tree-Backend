package handler

import (
	"context"
	"time"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetOneNode(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	userID := c.Query("user_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	node, err := h.nodeSvc.GetNodeDetails(ctx, nodeID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "retrieved node details successfully", "node_id", nodeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Node details retrieved successfully",
		"data":    node,
	})
}

func (h *Handler) CreateNode(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	var req model.CreateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.PathID = pathID

	h.logger.InfoContext(ctx, "adding node to path", "path_id", pathID, "node_title", req.Title)

	id, err := h.nodeSvc.AddNode(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "node added successfully", "path_id", pathID, "node_id", id)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Node added to learning path successfully",
		"data": fiber.Map{
			"node_id": id,
		},
	})
}

func (h *Handler) UpdateNode(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	var req model.UpdateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "updating node", "node_id", nodeID)

	if err := h.nodeSvc.EditNode(ctx, nodeID, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "node updated successfully", "node_id", nodeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Node details updated successfully",
		"data": fiber.Map{
			"node_id": nodeID,
		},
	})
}

func (h *Handler) DeleteNode(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "requesting node deletion", "node_id", nodeID)

	if err := h.nodeSvc.RemoveNode(ctx, nodeID); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "node deleted successfully", "node_id", nodeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Node has been deleted successfully",
		"data": fiber.Map{
			"node_id": nodeID,
		},
	})
}

func (h *Handler) CreateMaterial(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	var req model.CreateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.NodeID = nodeID

	h.logger.InfoContext(ctx, "adding material to node", "node_id", nodeID, "material_type", req.Type)

	id, err := h.nodeSvc.AddMaterial(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "material added successfully", "node_id", nodeID, "material_id", id)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Material added to node successfully",
		"data": fiber.Map{
			"material_id": id,
		},
	})
}

func (h *Handler) DeleteMaterial(c *fiber.Ctx) error {
	material_id := c.Params("material_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "deleting material", "material_id", material_id)

	if err := h.nodeSvc.RemoveMaterial(ctx, material_id); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "material deleted successfully", "material_id", material_id)
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Material has been deleted successfully",
		"data": fiber.Map{
			"material_id": material_id,
		},
	})
}

func (h *Handler) ReorderNodes(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	var req model.ReorderNodesRequest

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "reordering nodes in path", "path_id", pathID, "nodes_count", len(req.NodeIDs))

	if err := h.nodeSvc.ReorderNodes(ctx, pathID, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "nodes reordered successfully", "path_id", pathID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Nodes reordered successfully",
	})
}

func (h *Handler) GetNodesByPathID(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "fetching nodes for path", "path_id", pathID)

	nodes, err := h.nodeSvc.GetNodesByPathID(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Nodes retrieved successfully",
		"data":    nodes,
	})
}
