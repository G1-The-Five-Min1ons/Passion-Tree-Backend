package handler

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"
)

func (h *Handler) Create(c *fiber.Ctx) error {
	var req model.CreateReflectionRequest

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	res, err := h.service.CreateReflection(c.Context(), req)
	if err != nil {
		return h.handleError(c, apperror.NewInternal(err))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "reflection created successfully",
		"data":    res,
	})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("reflect_id")

	var req model.UpdateReflectionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.service.UpdateReflection(c.Context(), id, req); err != nil {
		return h.handleError(c, apperror.NewInternal(err))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "reflection updated successfully",
	})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("reflect_id")

	if err := h.service.DeleteReflection(c.Context(), id); err != nil {
		return h.handleError(c, apperror.NewInternal(err))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "reflection deleted successfully",
	})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("reflect_id")

	res, err := h.service.GetReflectionByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, apperror.NewInternal(err))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": res,
	})
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	res, err := h.service.GetAllReflections(c.Context())
	if err != nil {
		return h.handleError(c, apperror.NewInternal(err))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": res,
	})
}
