package handler

import (
	"context"
	"time"
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
	
	"github.com/gofiber/fiber/v2"
)

// Register creates a new user with profile
func (h *Handler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// กำหนดค่าเริ่มต้นและป้องกัน Privilege Escalation โดย Hard-code Role เป็น user
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user",
	}

	profile := &model.Profile{
		Bio:       req.Bio,
		Location:  req.Location,
		AvatarURL: req.AvatarURL,
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := h.userSvc.CreateUser(ctx, user, profile)
	if err != nil {
		return h.handleError(c, err)
	}

	// Auto-login หลังจากสมัครสมาชิกสำเร็จ
	token, _ := h.userSvc.Login(ctx, req.Username, req.Password)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
		"data": fiber.Map{
			"user_id": userID,
			"token":   token,
		},
	})
}

// Login authenticates a user
func (h *Handler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	token, err := h.userSvc.Login(ctx, req.Identifier, req.Password)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"token": token},
	})
}

// GetUserProfile gets user and profile by ID from JWT token
func (h *Handler) GetUserProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	user, profile, err := h.userSvc.GetUserByID(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User profile retrieved successfully",
		"data": fiber.Map{
			"user":    user,
			"profile": profile,
		},
	})
}

// UpdateUser updates user information from JWT token (only first_name and last_name)
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.UpdateUser(ctx, userID, req.FirstName, req.LastName); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
		"data": fiber.Map{
			"user_id": userID,
		},
	})
}

// DeleteUser deletes a user from JWT token with password confirmation
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.DeleteUser(ctx, userID, req.Password); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
		"data": fiber.Map{
			"user_id": userID,
		},
	})
}
