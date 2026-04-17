package handler

import (
	"context"
	"strconv"
	"strings"
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

	// Validate Role
	if req.Role != model.RoleStudent && req.Role != model.RoleTeacher && req.Role != model.RolePending {
		return h.handleError(c, apperror.NewBadRequest("role must be either 'student', 'teacher', or 'pending'"))
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

	h.logger.InfoContext(ctx, "user registered successfully", "user_id", userID)

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

	// Check if user is confirming reactivation
	confirmReactivate := c.Query("confirm_reactivate", "false") == "true"

	accessToken, refreshToken, err := h.userSvc.Login(ctx, req.Identifier, req.Password, deviceInfo, ipAddress, userAgent)

	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			if appErr.Code == fiber.StatusForbidden && appErr.Type == "OTP_REQUIRED" {
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"success":      true,
					"requires_otp": true,
					"message":      appErr.Message,
				})
			}
			if appErr.Code == fiber.StatusForbidden && appErr.Type == "ACCOUNT_DEACTIVATED" {
				// User's account is deactivated
				if !confirmReactivate {
					// Ask for confirmation
					return c.Status(fiber.StatusOK).JSON(fiber.Map{
						"success":               true,
						"requires_reactivation": true,
						"message":               appErr.Message,
						"grace_period_days":     model.DefaultDeactivationGracePeriod,
					})
				}

				// User confirmed reactivation, clear deactivation and try login again
				// Get user ID to reactivate
				var user *model.User
				if strings.Contains(req.Identifier, "@") {
					user, err = h.userSvc.GetUserByEmail(ctx, req.Identifier)
				} else {
					user, err = h.userSvc.GetUserByUsername(ctx, req.Identifier)
				}
				if err != nil || user == nil {
					return h.handleError(c, apperror.NewUnauthorized("user not found"))
				}

				// Reactivate account
				if err := h.userSvc.ReactivateAccount(ctx, user.UserID); err != nil {
					return h.handleError(c, err)
				}

				// Try login again
				accessToken, refreshToken, err = h.userSvc.Login(ctx, req.Identifier, req.Password, deviceInfo, ipAddress, userAgent)
				if err != nil {
					return h.handleError(c, err)
				}

				h.logger.InfoContext(ctx, "user_login_after_reactivation", "user_id", user.UserID, "ip", ipAddress)
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"success": true,
					"data": fiber.Map{
						"access_token":  accessToken,
						"refresh_token": refreshToken,
					},
				})
			}
		}
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "user_login_success", "ip", ipAddress)

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

	// Check if account is deactivated
	deactivatedUntil, err := h.userSvc.GetAccountDeactivatedUntil(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to check account deactivation status", "error", err, "user_id", userID)
		return h.handleError(c, apperror.NewInternal("failed to verify account status: %w", err))
	}

	// Account is deactivated and grace period not expired
	if deactivatedUntil != nil && deactivatedUntil.After(time.Now().UTC()) {
		return h.handleError(c, apperror.NewForbidden("account is deactivated. reactivate within the grace period to access your profile"))
	}

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

// UpdateUser updates user information from JWT token (first_name, last_name, and optionally role)
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return h.handleError(c, apperror.NewUnauthorized("please login again"))
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// Validate role if provided
	if req.Role != "" && req.Role != model.RoleStudent && req.Role != model.RoleTeacher {
		return h.handleError(c, apperror.NewBadRequest("role must be either 'student' or 'teacher'"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.UpdateUser(ctx, userID, "", req.FirstName, req.LastName, string(req.Role)); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "user updated successfully", "user_id", userID)

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
	if err != nil || userID == "" {
		return h.handleError(c, apperror.NewUnauthorized("invalid session"))
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

	h.logger.InfoContext(ctx, "user_deleted", "user_id", userID, "ip", c.IP())

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User account has been permanently deleted",
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

// Logout revokes one refresh token when provided, otherwise revokes all refresh tokens.
func (h *Handler) Logout(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return h.handleError(c, apperror.NewUnauthorized("invalid session"))
	}

	var req model.LogoutRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return h.handleError(c, apperror.NewBadRequest("invalid request body"))
		}
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	message := "Logged out successfully"
	if req.RefreshToken != "" {
		if err := h.userSvc.LogoutByRefreshToken(ctx, userID, req.RefreshToken); err != nil {
			return h.handleError(c, err)
		}
		message = "Session logged out successfully"
	} else {
		if err := h.userSvc.Logout(ctx, userID); err != nil {
			return h.handleError(c, err)
		}
	}

	h.logger.InfoContext(ctx, "user logged out successfully", "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
	})
}

// DeactivateAccount deactivates account temporarily and revokes active refresh sessions
func (h *Handler) DeactivateAccount(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return h.handleError(c, apperror.NewUnauthorized("invalid session"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.DeactivateAccount(ctx, userID, model.DefaultDeactivationGracePeriod); err != nil {
		h.logger.InfoContext(ctx, "account deactivation failed", "user_id", userID, "error", err)
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "account deactivated successfully", "user_id", userID)

	message := "User account deactivated for " + strconv.Itoa(model.DefaultDeactivationGracePeriod) + " days"

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data": fiber.Map{
			"user_id": userID,
		},
	})
}

// ReactivateAccount reactivates a temporary deactivated account
func (h *Handler) ReactivateAccount(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return h.handleError(c, apperror.NewUnauthorized("invalid session"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ReactivateAccount(ctx, userID); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "account reactivated successfully", "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Account reactivated successfully",
		"data": fiber.Map{
			"user_id": userID,
		},
	})
}

// GetActiveSessions retrieves all active sessions/devices for the authenticated user
func (h *Handler) GetActiveSessions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return h.handleError(c, apperror.NewUnauthorized("invalid session"))
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
	if err != nil || userID == "" {
		return h.handleError(c, apperror.NewUnauthorized("invalid session"))
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

	h.logger.InfoContext(ctx, "session logged out successfully", "user_id", userID, "session_id", sessionID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Session logged out successfully",
	})
}
