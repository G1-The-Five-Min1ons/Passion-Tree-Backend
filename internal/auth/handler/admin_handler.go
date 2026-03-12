package handler

import (
	"context"
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"time"

	"github.com/gofiber/fiber/v2"
)

type adminCreateUserRequest struct {
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Role      model.UserRole `json:"role"`
}

type adminUpdateUserRequest struct {
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Role      model.UserRole `json:"role"`
}

// GetAllUsers returns all users (admin only)
func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.userSvc.GetAllUsers(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get all users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve users",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Users retrieved successfully",
		"data":    users,
	})
}

// GetDashboardStats returns dashboard statistics (admin only)
func (h *Handler) GetDashboardStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.userSvc.GetDashboardStats(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get dashboard stats", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve dashboard stats",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Dashboard stats retrieved successfully",
		"data":    stats,
	})
}

func (h *Handler) CreateUserByAdmin(c *fiber.Ctx) error {
	var req adminCreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if req.Role == "" {
		req.Role = model.RoleStudent
	}

	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}

	profile := &model.Profile{}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := h.userSvc.CreateUserByAdmin(ctx, user, profile)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data": fiber.Map{
			"user_id": userID,
		},
	})
}

func (h *Handler) UpdateUserByAdmin(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	var req adminUpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.UpdateUserByAdmin(ctx, userID, req.FirstName, req.LastName, string(req.Role)); err != nil {
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

func (h *Handler) DeleteUserByAdmin(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.DeleteUserByAdmin(ctx, userID); err != nil {
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
