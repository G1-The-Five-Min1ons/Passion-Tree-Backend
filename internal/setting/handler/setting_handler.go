package handler

import (
	"context"
	"time"

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

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	settings, err := h.svc.GetSettings(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully retrieved settings", "user_id", userID, "count", len(settings))

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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	
	setting, err := h.svc.GetSetting(ctx, userID, key)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully retrieved setting", "user_id", userID, "key", key)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Setting retrieved successfully",
		"data":    setting,
	})
}

// UpdateSettings updates multiple settings at once in a single atomic transaction
func (h *Handler) UpdateSettings(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var requests []model.SettingRequest
	if err := c.BodyParser(&requests); err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	// Use batch update method for atomicity and performance
	if err := h.svc.UpdateMultipleSettings(ctx, userID, requests); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully updated settings", "user_id", userID, "count", len(requests))

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
		return h.handleError(c, err)
	}

	req.Key = key

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.svc.UpdateSetting(ctx, userID, &req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully updated setting", "user_id", userID, "key", key)

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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.svc.DeleteSetting(ctx, userID, key); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully deleted setting", "user_id", userID, "key", key)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Setting deleted successfully",
		"data": fiber.Map{
			"key": key,
		},
	})
}
