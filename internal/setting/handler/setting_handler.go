package handler

import (
	"passiontree/internal/pkg/middleware"
	"passiontree/internal/setting/model"

	"github.com/gofiber/fiber/v2"
)

// GetSettings retrieves all settings for the authenticated user
func (h *Handler) GetSettings(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	settings, err := h.svc.GetSettings(c.Context(), userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(c.Context(), "successfully retrieved settings", "user_id", userID, "count", len(settings))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Settings retrieved successfully",
		"data":    settings,
	})
}

// GetSetting retrieves a specific setting by key
func (h *Handler) GetSetting(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	key := c.Params("key")
	setting, err := h.svc.GetSetting(c.Context(), userID, key)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Setting retrieved successfully",
		"data":    setting,
	})
}

// UpdateSettings updates multiple settings at once
func (h *Handler) UpdateSettings(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var requests []model.SettingRequest
	if err := c.BodyParser(&requests); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	for _, req := range requests {
		if err := h.svc.UpdateSetting(c.Context(), userID, &req); err != nil {
			return h.handleError(c, err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Settings updated successfully",
		"data":    requests,
	})
}

// UpdateSetting updates a specific setting by key
func (h *Handler) UpdateSetting(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	key := c.Params("key")

	var req model.SettingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.Key = key

	if err := h.svc.UpdateSetting(c.Context(), userID, &req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Setting updated successfully",
		"data":    req,
	})
}

// DeleteSetting deletes a specific setting by key
func (h *Handler) DeleteSetting(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	key := c.Params("key")
	if err := h.svc.DeleteSetting(c.Context(), userID, key); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Setting deleted successfully",
		"data": fiber.Map{
			"key": key,
		},
	})
}
