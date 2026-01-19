package handler

import (
	"context"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

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

func (h *Handler) GetPathProgress(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	progress, err := h.pathSvc.GetPathProgress(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path progress calculated successfully",
		"data":    progress,
	})
}
func (h *Handler) Generate(c *fiber.Ctx) error {
	var req model.AIGeneratePathRequest

	ctx, cancel := context.WithTimeout(c.UserContext(), 60*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	result, err := h.pathSvc.GeneratePathWithAI(ctx, req.Topic)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Path generated successfully",
		"data":    result,
	})
}
