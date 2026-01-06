package handler

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

func (h *ReflectionHandler) Create(c *fiber.Ctx) error {
	var req model.CreateReflectionRequest

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	res, err := h.reflectSvc.CreateReflection(c.Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "reflection created successfully",
		"data": fiber.Map{
			"reflect_id": res,
		},
	})
}

func (h *ReflectionHandler) Update(c *fiber.Ctx) error {
	id := c.Params("reflect_id")

	var req model.UpdateReflectionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.reflectSvc.UpdateReflection(c.Context(), id, req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection updated successfully",
		"data": fiber.Map{
			"reflect_id": id,
		},
	})
}

func (h *ReflectionHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("reflect_id")

	if err := h.reflectSvc.DeleteReflection(c.Context(), id); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection deleted successfully",
		"data": fiber.Map{
			"reflect_id": id,
		},
	})
}

func (h *ReflectionHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("reflect_id")

	res, err := h.reflectSvc.GetReflectionByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection retrieved successfully",
		"data": fiber.Map{
			"reflect_id": res,
		},
	})
}

func (h *ReflectionHandler) GetAll(c *fiber.Ctx) error {
	res, err := h.reflectSvc.GetAllReflections(c.Context())
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflections retrieved successfully",
		"data": res,
	})
}
