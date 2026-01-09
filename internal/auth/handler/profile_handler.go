package handler

import (
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

// UpdateProfile updates user profile information
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	var profile model.Profile

	if err := c.BodyParser(&profile); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	profile.UserID = userID
	if err := h.userSvc.UpdateProfile(userID, &profile); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data": fiber.Map{
			"user_id": userID,
		},
	})
}

// GetProfile gets profile by user ID
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	_, profile, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		return h.handleError(c, err)
	}

	if profile == nil {
		return h.handleError(c, apperror.NewNotFound("profile not found"))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile retrieved successfully",
		"data":    profile,
	})
}