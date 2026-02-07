package handler

import (
	"context"
	"time"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) Create(c *fiber.Ctx) error {
	var req model.CreateReflectionRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	res, err := h.reflectSvc.CreateReflection(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection created successfully", "reflect_id", res)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "reflection created successfully",
		"data":    res,
	})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("reflect_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateReflectionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.reflectSvc.UpdateReflection(ctx, id, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection updated successfully", "reflect_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection updated successfully",
		"data": fiber.Map{
			"reflect_id": id,
		},
	})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("reflect_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.reflectSvc.DeleteReflection(ctx, id); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection deleted successfully", "reflect_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection deleted successfully",
		"data": fiber.Map{
			"reflect_id": id,
		},
	})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("reflect_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	res, err := h.reflectSvc.GetReflectionByID(ctx, id)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection retrieved successfully", "reflect_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection retrieved successfully",
		"data": res,
	})
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	res, err := h.reflectSvc.GetAllReflections(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "all reflections retrieved", "count", len(res))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflections retrieved successfully",
		"data":    res,
	})
}
