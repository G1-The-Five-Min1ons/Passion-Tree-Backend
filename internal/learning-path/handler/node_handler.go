package handler

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (h *Handler) CreateNode(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	var req model.CreateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.PathID = pathID

	id, err := h.nodeSvc.AddNode(req)
	if err != nil {
		return h.handleError(c, err)
	}

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
	var req model.UpdateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.nodeSvc.EditNode(nodeID, req); err != nil {
		return h.handleError(c, err)
	}

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
	if err := h.nodeSvc.RemoveNode(nodeID); err != nil {
		return h.handleError(c, err)
	}
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
	var req model.CreateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.NodeID = nodeID

	id, err := h.nodeSvc.AddMaterial(req)
	if err != nil {
		return h.handleError(c, err)
	}

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
	if err := h.nodeSvc.RemoveMaterial(material_id); err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Material has been deleted successfully",
		"data": fiber.Map{
			"material_id": material_id,
		},
	})
}