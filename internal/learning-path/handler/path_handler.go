package handler

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (h *Handler) GetAll(c *fiber.Ctx) error {
	paths, err := h.pathSvc.GetPaths()
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": paths})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req model.CreatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	id, err := h.pathSvc.CreatePath(req)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "created", "path_id": id})
}

func (h *Handler) GetOne(c *fiber.Ctx) error {
	id := c.Params("path_id")
	path, err := h.pathSvc.GetPathDetails(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(path)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("path_id")
	var req model.UpdatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.pathSvc.UpdatePath(id, req); err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "updated"})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("path_id")
	if err := h.pathSvc.DeletePath(id); err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "deleted"})
}

func (h *Handler) Start(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	var req model.StartPathRequest
	if err := c.BodyParser(&req); err != nil {
		return handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.pathSvc.StartPath(pathID, req.UserID); err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "enrolled successfully"})
}

func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")

	if userID == "" {
		return handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	status, err := h.pathSvc.GetEnrollmentStatus(pathID, userID)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": status})
}