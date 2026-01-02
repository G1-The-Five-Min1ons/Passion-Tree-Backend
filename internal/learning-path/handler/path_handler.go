package handler

import (
    "github.com/gofiber/fiber/v2"
    "passiontree/internal/learning-path/model"
)

func (h *Handler) GetAll(c *fiber.Ctx) error {
	paths, err := h.pathSvc.GetPaths()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": paths})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req model.CreatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	id, err := h.pathSvc.CreatePath(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "created", "path_id": id})
}

func (h *Handler) GetOne(c *fiber.Ctx) error {
	id := c.Params("path_id")
	path, err := h.pathSvc.GetPathDetails(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "path not found or error fetching data"})
	}
	return c.Status(fiber.StatusOK).JSON(path)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("path_id")
	var req model.UpdatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.pathSvc.UpdatePath(id, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "updated"})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("path_id")
	if err := h.pathSvc.DeletePath(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "deleted"})
}

func (h *Handler) Start(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	var req model.StartPathRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.pathSvc.StartPath(pathID, req.UserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "enrolled successfully"})
}

func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	status, err := h.pathSvc.GetEnrollmentStatus(pathID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": status})
}