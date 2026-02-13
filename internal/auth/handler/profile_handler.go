package handler

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// UpdateProfile updates user profile information
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
        return h.handleError(c, err)
    }

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	profileData := &model.Profile{
		UserID:    userID,
		AvatarURL: req.AvatarURL,
		Location:  req.Location,
		Bio:       req.Bio,
	}

	if err := h.userSvc.UpdateProfile(ctx, userID, profileData); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "profile updated successfully", "user_id", userID, "client_ips", c.IP())

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data": fiber.Map{
			"user_id": userID,
			"profile": profileData,
		},
	})
}

// GetProfile gets profile by user ID
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
    if err != nil {
        return h.handleError(c, err)
    }
	
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	_, profile, err := h.userSvc.GetUserByID(ctx, userID)
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
