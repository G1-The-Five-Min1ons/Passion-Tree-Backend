package handler

import (
	"context"
	"time"

	"passiontree/internal/mission/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) CreateTemplate(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	id, err := h.missionSvc.CreateTemplate(ctx, req, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Mission template created successfully",
		"data":    fiber.Map{"mission_id": id},
	})
}

func (h *Handler) GetMyMissions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	missions, err := h.missionSvc.GetMyMissions(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Retrieved active missions successfully",
		"data":    missions,
	})
}

func (h *Handler) TriggerAutoAssign(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 60*time.Second)
	defer cancel()

	if err := h.missionSvc.AutoAssignWeeklyMissions(ctx); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Weekly missions auto-assigned triggered successfully",
	})
}
