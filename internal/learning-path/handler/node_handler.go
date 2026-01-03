package handler

import (
    "github.com/gofiber/fiber/v2"
    "passiontree/internal/learning-path/model"
)

func (h *Handler) CreateNode(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	var req model.CreateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.PathID = pathID

	id, err := h.nodeSvc.AddNode(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"node_id": id})
}

func (h *Handler) UpdateNode(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req model.UpdateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.nodeSvc.EditNode(nodeID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "node updated"})
}

func (h *Handler) DeleteNode(c *fiber.Ctx) error {
	if err := h.nodeSvc.RemoveNode(c.Params("node_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "node deleted"})
}

func (h *Handler) CreateMaterial(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req model.CreateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.NodeID = nodeID

	id, err := h.nodeSvc.AddMaterial(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"material_id": id})
}

func (h *Handler) DeleteMaterial(c *fiber.Ctx) error {
	if err := h.nodeSvc.RemoveMaterial(c.Params("material_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "material deleted"})
}