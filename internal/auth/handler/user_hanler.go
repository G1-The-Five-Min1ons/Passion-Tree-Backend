package handler

import (
	"context"
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Register creates a new user with profile
func (h *Handler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// Validate Role
	if req.Role != model.RoleStudent && req.Role != model.RoleTeacher {
		return h.handleError(c, apperror.NewBadRequest("role must be either 'student' or 'teacher'"))
	}

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

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := h.userSvc.CreateUser(ctx, user, profile)
	if err != nil {
		return h.handleError(c, err)
	}

	// Extract device info for token tracking
	deviceInfo := c.Get("User-Agent", "Unknown Device")
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "Unknown")

	// Auto-login หลังจากสมัครสมาชิกสำเร็จ
	accessToken, refreshToken, err := h.userSvc.Login(ctx, req.Username, req.Password, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
		"data": fiber.Map{
			"user_id":       userID,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
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

	// Extract device info for session tracking
	deviceInfo := c.Get("User-Agent", "Unknown Device")
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "Unknown")

	accessToken, refreshToken, err := h.userSvc.Login(ctx, req.Identifier, req.Password, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
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

// RefreshToken generates a new access token and refresh token using token rotation
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var req model.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if req.RefreshToken == "" {
		return h.handleError(c, apperror.NewBadRequest("refresh_token is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	// Extract device info for new token
	deviceInfo := c.Get("User-Agent", "Unknown Device")
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "Unknown")

	accessToken, newRefreshToken, err := h.userSvc.RefreshAccessToken(ctx, req.RefreshToken, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tokens refreshed successfully",
		"data": fiber.Map{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
		},
	})
}

// Logout revokes all refresh tokens for the authenticated user
func (h *Handler) Logout(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.Logout(ctx, userID); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetActiveSessions retrieves all active sessions/devices for the authenticated user
func (h *Handler) GetActiveSessions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	// Try to get current refresh token from request body (optional)
	var req struct {
		CurrentRefreshToken string `json:"current_refresh_token"`
	}
	_ = c.BodyParser(&req)

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	sessions, err := h.userSvc.GetActiveSessions(ctx, userID, req.CurrentRefreshToken)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Active sessions retrieved successfully",
		"data":    sessions,
	})
}

// LogoutSession revokes a specific session by session ID
func (h *Handler) LogoutSession(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	sessionID := c.Params("session_id")
	if sessionID == "" {
		return h.handleError(c, apperror.NewBadRequest("session_id is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.LogoutSession(ctx, userID, sessionID); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Session logged out successfully",
	})
}
