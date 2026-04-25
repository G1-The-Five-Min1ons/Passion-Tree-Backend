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
	clientRequestID := c.Get("X-Client-Request-Id")

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	startedAt := time.Now()

	h.logger.InfoContext(ctx, "mission fetch request received",
		"user_id", userID,
		"client_request_id", clientRequestID,
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"request_id", c.GetRespHeader("X-Request-ID"),
	)

	missions, err := h.missionSvc.GetMyMissions(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "mission fetch failed",
			"user_id", userID,
			"client_request_id", clientRequestID,
			"elapsed_ms", time.Since(startedAt).Milliseconds(),
			"error", err,
		)
		return h.handleError(c, err)
	}
	if clientRequestID != "" {
		c.Set("X-Client-Request-Id", clientRequestID)
	}

	h.logger.InfoContext(ctx, "mission fetch response sent",
		"user_id", userID,
		"client_request_id", clientRequestID,
		"missions_count", len(missions),
		"status", fiber.StatusOK,
		"elapsed_ms", time.Since(startedAt).Milliseconds(),
	)

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
