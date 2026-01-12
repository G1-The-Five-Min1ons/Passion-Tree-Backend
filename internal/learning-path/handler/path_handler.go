package handler

import (
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetAll(c *fiber.Ctx) error {
	ctx := c.Context()
	paths, err := h.pathSvc.GetPaths(ctx)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning paths retrieved all successfully",
		"data":    paths,
	})
}

func (h *Handler) GetOne(c *fiber.Ctx) error {
	id := c.Params("path_id")
	ctx := c.Context()
	path, err := h.pathSvc.GetPathDetails(ctx, id)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path retrieved by one successfully",
		"data":    path,
	})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req model.CreatePathRequest
	ctx := c.Context()
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	id, err := h.pathSvc.CreatePath(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Learning path created successfully",
		"data": fiber.Map{
			"path_id": id,
		},
	})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("path_id")
	ctx := c.Context()
	var req model.UpdatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.pathSvc.UpdatePath(ctx, id, req); err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path updated successfully",
		"data": fiber.Map{
			"path_id": id,
		},
	})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("path_id")
	ctx := c.Context()
	if err := h.pathSvc.DeletePath(ctx, id); err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path deleted successfully",
		"data": fiber.Map{
			"path_id": id,
		},
	})
}

func (h *Handler) Start(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	ctx := c.Context()
	var req model.StartPathRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.pathSvc.StartPath(ctx, pathID, req.UserID); err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User enrolled in learning path successfully",
		"data": fiber.Map{
			"user_id": req.UserID,
			"path_id": pathID,
		},
	})
}

func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")
	ctx := c.Context()

	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	status, err := h.pathSvc.GetEnrollmentStatus(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Enrollment status retrieved successfully",
		"data":    status,
	})
}
