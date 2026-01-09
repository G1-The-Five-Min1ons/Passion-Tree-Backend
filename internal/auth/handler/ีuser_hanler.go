package handler

import (
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

// Register creates a new user with profile
func (h *Handler) Register(c *fiber.Ctx) error {
	var req struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
		Bio       string `json:"bio"`
		Location  string `json:"location"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// Create user and profile from request
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}

	profile := &model.Profile{
		Bio:       req.Bio,
		Location:  req.Location,
		AvatarURL: req.AvatarURL,
	}

	userID, err := h.userSvc.CreateUser(user, profile)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
		"data": fiber.Map{
			"user_id": userID,
		},
	})
}

// Login authenticates a user
func (h *Handler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	token, err := h.userSvc.Login(req.Email, req.Password)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"data": fiber.Map{
			"token": token,
		},
	})
}

// GetUserProfile gets user and profile by ID
func (h *Handler) GetUserProfile(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	user, profile, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		return h.handleError(c, err)
	}

	if user == nil {
		return h.handleError(c, apperror.NewNotFound("user not found"))
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

// UpdateUser updates user information
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	var user model.User

	if err := c.BodyParser(&user); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.userSvc.UpdateUser(userID, &user); err != nil {
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

// DeleteUser deletes a user
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	if err := h.userSvc.DeleteUser(userID); err != nil {
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
