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

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with associated profile information.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterRequest  true  "Registration payload"
// @Success      201   {object}  apidoc.UserIDResponse
// @Failure      400   {object}  apidoc.ErrorResponse  "Invalid request body or invalid role"
// @Failure      409   {object}  apidoc.ErrorResponse  "Username or email already exists"
// @Failure      500   {object}  apidoc.ErrorResponse  "Internal server error"
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Authenticate a user
// @Description  Logs in with username/email + password and returns access & refresh tokens.
// @Description  May also return `requires_otp` or `requires_reactivation` flags instead of tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        confirm_reactivate  query     string                false  "Set to 'true' to confirm reactivation of a deactivated account"
// @Param        body                body      model.LoginRequest    true   "Login credentials"
// @Success      200                 {object}  apidoc.TokenPairResponse
// @Success      200                 {object}  apidoc.OTPRequiredResponse           "OTP verification required"
// @Success      200                 {object}  apidoc.ReactivationRequiredResponse  "Account deactivated, awaiting reactivation confirmation"
// @Failure      400                 {object}  apidoc.ErrorResponse
// @Failure      401                 {object}  apidoc.ErrorResponse  "Invalid credentials"
// @Failure      429                 {object}  apidoc.ErrorResponse  "Too many login attempts"
// @Router       /auth/login [post]
func (h *Handler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	// Normalize identifier so trailing spaces / casing don't cause intermittent
	// "user not found" responses. Email is case-insensitive; username keeps its
	// original case but with whitespace trimmed.
	req.Identifier = strings.TrimSpace(req.Identifier)
	if strings.Contains(req.Identifier, "@") {
		req.Identifier = strings.ToLower(req.Identifier)
	}

	// Login does both DB lookups and an outbound Gmail SMTP call (which itself
	// has a 10s timeout). Give the whole flow more headroom so a slow first
	// SMTP handshake doesn't surface to the user as "user not found".
	ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
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

// GetUserProfile godoc
// @Summary      Get authenticated user's profile
// @Description  Returns the user object and profile of the JWT-authenticated user.
// @Tags         Profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse  "Account deactivated"
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /auth/profile [get]
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

// UpdateUser godoc
// @Summary      Update authenticated user's basic info
// @Description  Updates first name, last name, and optionally role for the JWT-authenticated user.
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.UpdateUserRequest  true  "Fields to update"
// @Success      200   {object}  apidoc.UserIDResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Failure      500   {object}  apidoc.ErrorResponse
// @Router       /auth/user [put]
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

// DeleteUser godoc
// @Summary      Permanently delete the authenticated user
// @Description  Requires password confirmation in the request body. This action is irreversible.
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      object{password=string}  true  "Password confirmation"
// @Success      200   {object}  apidoc.UserIDResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Failure      500   {object}  apidoc.ErrorResponse
// @Router       /auth/user [delete]
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

// RefreshToken godoc
// @Summary      Rotate access & refresh tokens
// @Description  Exchanges a valid refresh token for a new access token and a freshly rotated refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RefreshTokenRequest  true  "Refresh token payload"
// @Success      200   {object}  apidoc.TokenPairResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse  "Refresh token expired or revoked"
// @Router       /auth/refresh [post]
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

// Logout godoc
// @Summary      Logout (revoke refresh tokens)
// @Description  If a refresh_token is provided in the body, only that session is revoked.
// @Description  Otherwise all refresh tokens for the user are revoked.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.LogoutRequest  false  "Optional specific refresh token to revoke"
// @Success      200   {object}  apidoc.MessageResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /auth/logout [post]
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

	var logoutErr error
	finalMessage := "Logged out successfully"
	logoutType := "all_sessions"

	if req.RefreshToken != "" {
		logoutErr = h.userSvc.LogoutByRefreshToken(ctx, userID, req.RefreshToken)
		finalMessage = "Specific session revoked successfully"
		logoutType = "specific_session"
	} else {
		logoutErr = h.userSvc.Logout(ctx, userID)
		finalMessage = "All sessions logged out successfully"
	}

	if logoutErr != nil {
		if !apperror.IsNotFound(logoutErr) {
			return h.handleError(c, logoutErr)
		}
	}

	h.logger.InfoContext(ctx, "user logout activity recorded", "user_id", userID, "type", logoutType)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": finalMessage,
	})
}

// DeactivateAccount godoc
// @Summary      Temporarily deactivate the authenticated user's account
// @Description  Marks the account as deactivated for the configured grace period and revokes all active refresh sessions.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.UserIDResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /auth/deactivate [post]
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

// ReactivateAccount godoc
// @Summary      Reactivate a deactivated account
// @Description  Clears the deactivation flag for the JWT-authenticated user, restoring access immediately.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.UserIDResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /auth/reactivate [post]
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

// GetActiveSessions godoc
// @Summary      List active sessions / devices
// @Description  Returns the list of active refresh-token sessions for the authenticated user.
// @Description  Pass the current refresh token in the body to flag the current session in the response.
// @Tags         Sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      object{current_refresh_token=string}  false  "Optional current refresh token"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /auth/sessions [get]
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

// LogoutSession godoc
// @Summary      Logout a specific session/device
// @Description  Revokes the refresh token associated with the given session ID.
// @Tags         Sessions
// @Produce      json
// @Security     BearerAuth
// @Param        session_id  path      string  true  "Session ID to revoke"
// @Success      200         {object}  apidoc.MessageResponse
// @Failure      400         {object}  apidoc.ErrorResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Router       /auth/sessions/{session_id} [delete]
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
