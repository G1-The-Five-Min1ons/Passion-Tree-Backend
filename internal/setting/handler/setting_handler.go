package handler

import (
	"context"
	"time"

	"passiontree/internal/pkg/middleware"
	"passiontree/internal/setting/model"

	"github.com/gofiber/fiber/v2"
)

// GetSettings godoc
// @Summary      List all settings for the authenticated user
// @Tags         Settings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Router       /settings/ [get]
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

// GetSetting godoc
// @Summary      Get a specific setting by key
// @Tags         Settings
// @Produce      json
// @Security     BearerAuth
// @Param        key  path      string  true  "Setting key"
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      404  {object}  apidoc.ErrorResponse
// @Router       /settings/{key} [get]
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

// UpdateSettings godoc
// @Summary      Bulk update settings (atomic transaction)
// @Description  Accepts an array of {key, value} entries and persists them in one transaction.
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      []model.SettingRequest  true  "Array of setting requests"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /settings/ [put]
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

// UpdateSetting godoc
// @Summary      Update a single setting by key
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        key   path      string                true  "Setting key"
// @Param        body  body      model.SettingRequest  true  "Setting payload"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /settings/{key} [put]
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

// DeleteSetting godoc
// @Summary      Delete a setting by key
// @Tags         Settings
// @Produce      json
// @Security     BearerAuth
// @Param        key  path      string  true  "Setting key"
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      404  {object}  apidoc.ErrorResponse
// @Router       /settings/{key} [delete]
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
