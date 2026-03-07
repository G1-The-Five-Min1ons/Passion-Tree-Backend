package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
)

func (h *Handler) ListTeacherApplications(c *fiber.Ctx) error {
	status := c.Query("status")

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	applications, err := h.userSvc.GetTeacherApplications(ctx, status)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Teacher applications retrieved successfully",
		"data":    applications,
	})
}

func (h *Handler) ReviewTeacherApplication(c *fiber.Ctx) error {
	adminID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	requestID := c.Params("request_id")
	if requestID == "" {
		return h.handleError(c, apperror.NewBadRequest("request_id is required"))
	}

	var req model.ReviewTeacherApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ReviewTeacherApplication(ctx, requestID, adminID, req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Teacher application reviewed successfully",
	})
}
