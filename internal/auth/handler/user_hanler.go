package handler

import (
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

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

	// Auto-login: Generate token after registration
	user.UserID = userID // Set the generated user ID
	token, err := h.userSvc.Login(user.Username, req.Password)
	if err != nil {
		// Registration succeeded but auto-login failed, return without token
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "User registered successfully. Please login.",
			"data": fiber.Map{
				"user_id": userID,
			},
		})
	}

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
	var req struct {
		Identifier string `json:"identifier" binding:"required"` // Can be username or email
		Password   string `json:"password" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	token, err := h.userSvc.Login(req.Identifier, req.Password)
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

// GetUserProfile gets user and profile by ID from JWT token
func (h *Handler) GetUserProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

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

// UpdateUser updates user information from JWT token (only first_name and last_name)
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var user model.User

	if err := c.BodyParser(&user); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// Only allow updating first_name and last_name
	if err := h.userSvc.UpdateUser(userID, user.FirstName, user.LastName); err != nil {
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

	if err := h.userSvc.DeleteUser(userID, req.Password); err != nil {
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

// VerifyEmail verifies user's email with verification code
func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if req.Code == "" {
		return h.handleError(c, apperror.NewBadRequest("verification code is required"))
	}

	if err := h.userSvc.VerifyEmail(req.Code); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Email verified successfully",
	})
}

// ResendVerificationEmail resends verification email
func (h *Handler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.userSvc.ResendVerificationEmail(req.Email); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Verification email sent successfully",
	})
}

// ForgotPassword sends password reset code
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.userSvc.ForgotPassword(req.Email); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "If the email exists, a password reset code has been sent",
	})
}

// ResetPassword resets password using code
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.userSvc.ResetPassword(req.Code, req.NewPassword); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Password reset successfully",
	})
}

// ChangePassword changes password for authenticated user
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.userSvc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Password changed successfully",
	})
}
